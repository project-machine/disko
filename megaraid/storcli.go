package megaraid

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/patrickmn/go-cache"
	"machinerun.io/disko"
)

type storCli struct {
}

// StorCli returns a storcli specific implementation of Query
func StorCli() MegaRaid {
	return &storCli{}
}

type scResultSectionType int

const (
	// UnknownMedia - indicates an unknown media
	rsUnknown scResultSectionType = iota
	rsHeader
	rsVirtDisk
	rsPhysDisks
	rsVirtProps
	rsVdList
	rsDgDriveList
)

const noStorCliRC = 127

type scResultSection struct {
	Type  scResultSectionType
	Name  string
	Lines []string
}

func (sc *storCli) Query(cID int) (Controller, error) {
	// run /c0/dall show all
	//   - get PDs and VDs
	// run /c0/vall show all
	//   - populate VD Properties and Path
	var stdout, stderr []byte
	var rc int

	args := []string{fmt.Sprintf("/c%d", cID), "show", "nolog"}

	if stdout, stderr, rc = storcli(args...); rc != 0 {
		var err error = ErrNoStorcli
		if rc != noStorCliRC {
			err = cmdError(args, stdout, stderr, rc)
		}

		return Controller{}, err
	}

	cxDxOut := string(stdout)

	args = []string{fmt.Sprintf("/c%d/vall", cID), "show", "all", "nolog"}

	if stdout, stderr, rc = storcli(args...); rc != 0 {
		return Controller{}, cmdError(args, stdout, stderr, rc)
	}

	cxVxOut := string(stdout)

	return newController(cID, cxDxOut, cxVxOut)
}

func (sc *storCli) DriverSysfsPath() string {
	return SysfsPCIDriversPath
}

func (sc *storCli) GetDiskType(path string) (disko.DiskType, error) {
	return disko.HDD, fmt.Errorf("missing controller to run query")
}

// not implemented in driver layer
func (sc *storCli) IsSysPathRAID(syspath string) bool {
	return false
}

func newController(cID int, cxDxOut string, cxVxOut string) (Controller, error) {
	const pathPropName = "OS Drive Name"

	ctrl := Controller{
		ID: cID,
	}

	vds, pds, err := parseCxShow(cxDxOut)
	if err != nil {
		return ctrl, err
	}

	propMap, err := parseVirtProperties(cxVxOut)
	if err == ErrUnsupported {
		propMap = map[int](map[string]string){}
	} else if err != nil {
		return ctrl, err
	}

	ctrl.VirtDrives = vds
	ctrl.Drives = pds
	ctrl.DriveGroups = DriveGroupSet{}

	for vID, vProps := range propMap {
		ctrl.VirtDrives[vID].Properties = vProps
		ctrl.VirtDrives[vID].Path = vProps[pathPropName]
	}

	for diskID, drive := range pds {
		dgID := drive.DriveGroup
		if dgID < 0 {
			continue
		}

		dg, ok := ctrl.DriveGroups[dgID]

		if ok {
			dg.Drives[diskID] = drive
		} else {
			ctrl.DriveGroups[dgID] = &DriveGroup{
				ID:     dgID,
				Drives: DriveSet{diskID: drive},
			}
		}
	}

	// fmt.Printf("ctrl: %#v\n", propMap)
	// fmt.Printf("ctrl: %#v\n", ctrl)

	return ctrl, nil
}

// loadSections - parse a storcli output into sections.
func loadSections(cmdOut string) []scResultSection {
	var header = false
	var curSect scResultSection
	var last string

	equalLine := regexp.MustCompile("^[=]+$")
	rSects := []scResultSection{}

	matchers := []struct {
		stype scResultSectionType
		regex *regexp.Regexp
	}{
		// /c0/v1 :
		{rsVirtDisk, regexp.MustCompile("^/c[0-9]+/v[0-9]+ :$")},
		// PDs for VD 0
		{rsPhysDisks, regexp.MustCompile("^PDs for VD [0-9]+ :$")},
		// PD LIST (storcli /c0 show)
		{rsPhysDisks, regexp.MustCompile("^PD LIST :$")},
		// VD0 Properties (storcli /c0/vall show all)
		{rsVirtProps, regexp.MustCompile("^.*VD[0-9]+ Properties :$")},
		// VD LIST (storcli /c0/dall show all)
		{rsVdList, regexp.MustCompile("^VD LIST :$")},
		// DG DRIVE LIST (storcli /c0/dall show all)
		{rsDgDriveList, regexp.MustCompile("(DG Drive LIST|UN-CONFIGURED DRIVE LIST) :$")},
		{rsUnknown, regexp.MustCompile("^.* :$")},
	}

	for _, cur := range strings.Split(cmdOut, "\n") {
		if !header {
			// header always first.
			header = true
			curSect = scResultSection{
				Type:  rsHeader,
				Name:  "header",
				Lines: []string{},
			}
		} else if equalLine.MatchString(cur) {
			newType := rsUnknown
			for _, m := range matchers {
				if m.regex.MatchString(last) {
					newType = m.stype
					break
				}
			}
			// drop the trailing " :"
			name := last[:len(last)-2]
			rSects = append(rSects, curSect)
			curSect = scResultSection{
				Type:  newType,
				Name:  name,
				Lines: []string{},
			}
			cur = ""
		} else if last != "" {
			curSect.Lines = append(curSect.Lines, last)
		}

		last = cur
	}

	if last != "" {
		curSect.Lines = append(curSect.Lines, last)
	}

	return append(rSects, curSect)
}

func parseKeyValData(lines []string) map[string]string {
	data := map[string]string{}
	const tokNum2 = 2

	for _, line := range lines {
		if line == "" {
			continue
		}

		toks := strings.SplitN(line, " = ", tokNum2)
		if len(toks) != tokNum2 {
			continue
		}

		data[toks[0]] = toks[1]
	}

	return data
}

func filterTableData(lines []string) []string {
	// find the dataLines.  dataLines[0] will be header, all others are rows.
	dashLine := regexp.MustCompile("^-+$")
	dataLines := []string{}

	for i, sepCount := 0, 0; i < len(lines) && sepCount < 3; i++ {
		if lines[i] == "" {
		} else if dashLine.MatchString(lines[i]) {
			sepCount++
		} else {
			dataLines = append(dataLines, lines[i])
		}
	}

	return dataLines
}

func parseTableData(lines []string) []map[string]string {
	// data looks like:
	//   --------------------
	//   field1 field2  field3
	//   --------------------
	//   record record2 record3
	//   ...
	//   --------------------
	const space = ' '

	type colCand struct {
		Left, Right int
	}

	colCands := []*colCand{}
	left := -1
	leadingSpace := true

	// find contiguous sets of whitespace (column Candidates) in the header Line.
	dataLines := filterTableData(lines)

	for i, curChar := range dataLines[0] {
		if curChar == space {
			if left < 0 && !leadingSpace {
				left = i
			}
		} else {
			leadingSpace = false
			if left >= 0 {
				colCands = append(colCands, &colCand{left, i - 1})
				left = -1
			}
		}
	}

	// walk through each columnCandidate range and
	// find the first column where all lines have a space there.
	var column int
	var cuts = []int{}

	for _, colCand := range colCands {
		success := false
		for column = colCand.Left; column < colCand.Right && !success; column++ {
			success = true

			for _, line := range dataLines[1:] {
				if line[column] != space {
					success = false
					break
				}
			}
		}

		cuts = append(cuts, column)
	}

	return cutTableLines(dataLines, cuts)
}

func cutTableLines(dataLines []string, cuts []int) []map[string]string {
	// now cut data lines into pieces at the columns in 'cuts'
	var data = []map[string]string{}
	var headers = []string{}
	var row []rune
	var from, i int

	trim := func(a []rune) string {
		return strings.Trim(string(a), " ")
	}

	row = []rune(dataLines[0])

	for from, i = 0, 0; i < len(cuts); i++ {
		headers = append(headers, trim(row[from:cuts[i]]))
		from = cuts[i]
	}

	headers = append(headers, trim(row[from:]))

	for _, line := range dataLines[1:] {
		rowData := map[string]string{}
		row = []rune(line)
		from = 0

		for i := 0; i < len(cuts); i++ {
			rowData[headers[i]] = trim(row[from:cuts[i]])
			from = cuts[i]
		}

		rowData[headers[len(cuts)]] = trim(row[from:])
		data = append(data, rowData)
	}

	return data
}

func isHeaderNotFound(header map[string]string) bool {
	return header["Status"] == "Failure" &&
		strings.Contains(header["Description"], "not found")
}

func isHeaderUnsupported(header map[string]string) bool {
	return header["Status"] == "Failure" &&
		strings.Contains(header["Description"], "Un-supported command")
}

func getHeaderError(header map[string]string) error {
	if header["Status"] == "Success" {
		return nil
	}

	if isHeaderNotFound(header) {
		return ErrNoController
	} else if isHeaderUnsupported(header) {
		return ErrUnsupported
	}

	if _, err := strconv.Atoi(header["Controller"]); err != nil {
		return fmt.Errorf("storcli controller in header not an int: %s", err)
	}

	return fmt.Errorf("storcli command returned status: %s", header["Status"])
}

// Parse the output of 'storcli /c0 show'
func parseCxShow(cmdOut string) (VirtDriveSet, DriveSet, error) {
	vds := VirtDriveSet{}
	pds := DriveSet{}

	sections := loadSections(cmdOut)

	for _, sect := range sections {
		// Missing Cases: rsDgDriveList, rsUnknown, rsVirtDisk, rsVirtProps
		//exhaustive:ignore
		switch sect.Type {
		case rsHeader:
			if err := getHeaderError(parseKeyValData(sect.Lines)); err != nil {
				return vds, pds, err
			}

		case rsVdList:
			data := parseTableData(sect.Lines)
			for _, vdData := range data {
				vd, err := vdDataToVirtDrive(vdData)
				if err != nil {
					return vds, pds, err
				}

				vds[vd.ID] = &vd
			}
		case rsPhysDisks:
			data := parseTableData(sect.Lines)
			for _, pdData := range data {
				pd, err := pdDataToDrive(pdData)
				if err != nil {
					return vds, pds, err
				}

				pds[pd.ID] = &pd
			}
		}
	}

	return vds, pds, nil
}

// parseVirtProperties - return properties map ("VD0 Properties") by VirtDrive ID
//
//	cmdOut is output of 'storcli /c0/vall show all'
func parseVirtProperties(cmdOut string) (map[int](map[string]string), error) {
	var vID int
	var err error
	const tokNum2 = 2

	nameMatch := regexp.MustCompile("^VD([0-9]+) Properties$")
	vdmap := map[int](map[string]string){}

	sections := loadSections(cmdOut)

	for _, sect := range sections {
		// Missing cases: rsDgDriveList, rsPhysDisks, rsUnknown, rsVdList, rsVirtDisk
		//exhaustive:ignore
		switch sect.Type {
		case rsHeader:
			if err := getHeaderError(parseKeyValData(sect.Lines)); err != nil {
				return vdmap, err
			}
		case rsVirtProps:
			// Extract the VirtDrive Number from the Name (VD0 Properties)
			toks := nameMatch.FindStringSubmatch(sect.Name)

			if len(toks) != tokNum2 {
				return vdmap, fmt.Errorf("failed parsing section '%s'", sect.Name)
			}

			if vID, err = strconv.Atoi(toks[1]); err != nil {
				return vdmap, fmt.Errorf("failed to get int from section '%s'", sect.Name)
			}

			vdmap[vID] = parseKeyValData(sect.Lines)
		}
	}

	return vdmap, nil
}

func parseIntOrDash(field string) (int, error) {
	if field == "-" {
		return -1, nil
	}

	return strconv.Atoi(field)
}

// vdDataToVirtDrive - take single data VD row (parseTableData(..)) return a VirtDrive
func vdDataToVirtDrive(data map[string]string) (VirtDrive, error) {
	var dg, vdNum int
	var err error

	nilVd := VirtDrive{}
	dgvd := strings.Split(data["DG/VD"], "/")

	if dg, err = parseIntOrDash(dgvd[0]); err != nil {
		return nilVd, fmt.Errorf("failed to get DriveGroup from %s: %s", dgvd, err)
	}

	if vdNum, err = parseIntOrDash(dgvd[1]); err != nil {
		return nilVd, fmt.Errorf("failed to get VirtDrive Number from %s: %s", dgvd, err)
	}

	return VirtDrive{
		ID:         vdNum,
		DriveGroup: dg,
		Path:       "",
		RaidName:   data["Name"],
		Type:       data["TYPE"],
		Raw:        data,
	}, nil
}

func parseDriveGroupVal(val string) (int, error) {
	known := map[string]int{
		"-": -1, // None
		"F": -2, // Foreign
	}

	if found, ok := known[val]; ok {
		return found, nil
	}

	return parseIntOrDash(val)
}

func pdDataToDrive(data map[string]string) (Drive, error) {
	var err error
	var dID, dg, eID, slot int
	const tokNum2 = 2

	if dID, err = parseIntOrDash(data["DID"]); err != nil {
		return Drive{}, err
	}

	if dg, err = parseDriveGroupVal(data["DG"]); err != nil {
		return Drive{}, err
	}

	toks := strings.SplitN(data["EID:Slt"], ":", tokNum2)
	if len(toks) != tokNum2 {
		return Drive{},
			fmt.Errorf(
				"splitting EID:Slt data '%s' on ':'' returned %d fields, expected 2",
				data["EID:Slt"], len(toks))
	}

	if eID, err = parseIntOrDash(toks[0]); err != nil {
		return Drive{}, err
	}

	if slot, err = parseIntOrDash(toks[1]); err != nil {
		return Drive{}, err
	}

	return Drive{
		ID:         dID,
		EID:        eID,
		Slot:       slot,
		DriveGroup: dg,
		State:      data["State"],
		MediaType:  stringToMediaType(data["Med"]),
		Model:      data["Model"],
		Raw:        data,
	}, nil
}

func stringToMediaType(mtypeStr string) MediaType {
	kmap := map[string]MediaType{
		"UNKNOWN": UnknownMedia,
		"HDD":     HDD,
		"SSD":     SSD,
	}
	if mtype, ok := kmap[mtypeStr]; ok {
		return mtype
	}

	return UnknownMedia
}

func storcli(args ...string) ([]byte, []byte, int) {
	cmd := exec.Command("storcli", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return stdout.Bytes(), stderr.Bytes(), getCommandErrorRCDefault(err, noStorCliRC)
}

func cmdError(args []string, out []byte, err []byte, rc int) error {
	if rc == 0 {
		return nil
	}

	return fmt.Errorf(
		"command failed [%d]:\n cmd: %v\n out:%s\n err:%s",
		rc, args, out, err)
}

func getCommandErrorRCDefault(err error, rcError int) int {
	if err == nil {
		return 0
	}

	exitError, ok := err.(*exec.ExitError)
	if ok {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}

	return rcError
}

type cachingStorCli struct {
	mr    MegaRaid
	cache *cache.Cache
}

// CachingStorCli - just a cache for a MegaRaid
func CachingStorCli() MegaRaid {
	const longTime = 5 * time.Minute

	return &cachingStorCli{
		mr:    &storCli{},
		cache: cache.New(longTime, longTime),
	}
}

func (csc *cachingStorCli) Query(cID int) (Controller, error) {
	type qresult struct {
		ctrl Controller
		err  error
	}

	cacheName := fmt.Sprintf("query-%d", cID)
	cached, found := csc.cache.Get(cacheName)

	if found {
		ret := cached.(qresult)
		return ret.ctrl, ret.err
	}

	ctrl, err := csc.mr.Query(cID)
	csc.cache.Set(cacheName, qresult{ctrl: ctrl, err: err}, cache.DefaultExpiration)

	return ctrl, err
}

func (csc *cachingStorCli) GetDiskType(path string) (disko.DiskType, error) {
	ctrl, err := csc.Query(0)
	if err == nil {
		for _, vd := range ctrl.VirtDrives {
			if vd.Path == path {
				if ctrl.DriveGroups[vd.DriveGroup].IsSSD() {
					return disko.SSD, nil
				}

				return disko.HDD, nil
			}
		}
	} else if err != ErrNoStorcli && err != ErrNoController && err != ErrUnsupported {
		return disko.HDD, err
	}

	return disko.HDD, fmt.Errorf("cannot determine disk type")
}

func (csc *cachingStorCli) DriverSysfsPath() string {
	return csc.mr.DriverSysfsPath()
}

// not implemented in the driver layer
func (csc *cachingStorCli) IsSysPathRAID(syspath string) bool {
	return false
}
