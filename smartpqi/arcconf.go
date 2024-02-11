package smartpqi

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"machinerun.io/disko"
)

const (
	noArcConfRC = 127
)

func parseLogicalDevices(rawData string) ([]LogicalDevice, error) {
	logDevs := []LogicalDevice{}

	devices := strings.Split(rawData, "\n\n\n")
	for _, device := range devices {
		ldStart := strings.Index(device, "Logical Device number")
		if ldStart >= 0 {
			ldRawData := strings.TrimSpace(device[ldStart:])

			// parse singular logical device
			logDev, err := parseLogicalDeviceString(ldRawData)
			if err != nil {
				return []LogicalDevice{}, fmt.Errorf("error parsing logical device from arcconf getconfig ld: %s", err)
			}

			logDevs = append(logDevs, logDev)
		}
	}
	return logDevs, nil
}

func parseLogicalDeviceString(rawData string) (LogicalDevice, error) {
	// list the logical device keys we're not parsing at this time
	logicalDeviceSkipKeys := map[string]bool{
		"Device ID":                              true,
		"Last Consistency Check Completion Time": true,
		"Last Consistency Check Duration":        true,
		"Status of Logical Device":               true,
		"Stripe-unit size":                       true,
		"Full Stripe Size":                       true,
		"Device Type":                            true, // Data]
		"Boot Type":                              true, // Primary and Secondary]
		"Heads":                                  true, // 255]
		"Sectors Per Track":                      true, // 32]
		"Cylinders":                              true, // 65535]
		"Caching":                                true, // Enabled]
		"Mount Points":                           true, // Not Mounted]
		"LD Acceleration Method":                 true, // Controller Cache]
		"SED Encryption":                         true, // Disabled]
		"Volume Unique Identifier":               true, // 600508B1001C6DB81E7099960E5B5796]
		"Consistency Check Status":               true, // Not Applicable]
	}
	ld := LogicalDevice{}

	if len(rawData) < 1 {
		return ld, fmt.Errorf("cannot parse an empty string")
	}

	// break data into first line which has the logical device number and
	// the remaining data is in mostly key : value pairs
	lineData := strings.SplitN(rawData, "\n", 2)
	if len(lineData) != 2 {
		return ld, fmt.Errorf("expected exactly 2 lines of data, found %d in %q", len(lineData), lineData)
	}

	// extract the LogicalDrive ID from the first token
	// Logical Device number N
	toks := strings.Split(strings.TrimSpace(lineData[0]), " ")
	if len(toks) != 4 {
		return ld, fmt.Errorf("expected 4 tokens in %q, found %d", lineData[0], len(toks))
	}

	ldID, err := strconv.Atoi(toks[len(toks)-1])
	if err != nil {
		return LogicalDevice{}, fmt.Errorf("error while parsing integer from %q: %s", toks[len(toks)-1], err)
	}

	ld.ID = ldID

	// split the remainder into line data, split on :, but expect 2 tokens
	// since some lines have ':' in the value portion
	for _, lineRaw := range strings.Split(lineData[1], "\n") {
		// remove leading space formatting
		line := strings.TrimSpace(lineRaw)

		// ignore lines that didn't have colon in it
		rawToks := strings.SplitN(line, ":", 2)
		if len(rawToks) < 2 {
			continue
		}

		// the raw tokens split on colon will have whitespace, so let's trim
		// that into our final key, value tokens:
		// 'Logical Device Name               ',  'LogicalDrv 0'  =>  'Logical Device Name', 'LogicalDrv 0'
		toks := []string{}
		for _, tok := range rawToks {
			toks = append(toks, strings.TrimSpace(tok))
		}

		// skip tokens if key is on the ignore list
		if _, ok := logicalDeviceSkipKeys[toks[0]]; ok {
			continue
		}

		// map the key to LogicalDevice member field
		switch {
		case toks[0] == "Logical Device name": // LogicalDrv 0]
			ld.Name = toks[1]

		case toks[0] == "Disk Name": // /dev/sdc (Disk0) (Bus: 1, Target: 0, Lun: 2)]
			ld.DiskName = strings.Fields(toks[1])[0]

		case toks[0] == "Block Size of member drives": // 512 Bytes]
			bs, err := strconv.Atoi(strings.Fields(toks[1])[0])
			if err != nil {
				return ld, fmt.Errorf("failed to parse BlockSize from token '%s': %s", toks[1], err)
			}

			ld.BlockSize = bs
		case toks[0] == "Array": // 2]
			aID, err := strconv.Atoi(toks[1])
			if err != nil {
				return ld, fmt.Errorf("failed to parse ArrayID from token '%s': %s", toks[1], err)
			}

			ld.ArrayID = aID
		case toks[0] == "RAID level": // 0]
			// we don't parse RAIDLevel as integer since arrconf device support
			// non-numeric values like: 0, 1, 1Triple, 10, 10Triple, 5, 6, 50 and 60
			ld.RAIDLevel = toks[1]

		case toks[0] == "Size": // 1144609 MB]
			sizeMB, err := strconv.Atoi(strings.Fields(toks[1])[0])
			if err != nil {
				return ld, fmt.Errorf("failed to parse Size from token '%s': %s", toks[1], err)
			}

			ld.SizeMB = sizeMB
		case toks[0] == "Interface Type": // Serial Attached SCSI]
			switch toks[1] {
			case "Serial Attached SCSI":
				ld.InterfaceType = "SCSI"
			case "Serial Attached ATA":
				ld.InterfaceType = "ATA"
			}
		}
	}

	return ld, nil
}

func parsePhysicalDevices(output string) ([]PhysicalDevice, error) {
	// list the physical device keys we're not parsing
	physicalDeviceParseKeys := map[string]bool{
		"Array":               true,
		"Block Size":          true,
		"Firmware":            true,
		"Model":               true,
		"Physical Block Size": true,
		"Serial number":       true,
		"SSD":                 true,
		"State":               true,
		"Total Size":          true,
		"Vendor":              true,
		"Write Cache":         true,
	}
	pDevs := []PhysicalDevice{}
	deviceStartRe := regexp.MustCompile(`Device\ #\d+`)
	devStartIdx := []int{}
	devEndIdx := []int{}

	devStart := deviceStartRe.FindAllIndex([]byte(output), -1)
	if len(devStart) < 1 {
		return []PhysicalDevice{}, fmt.Errorf("error finding start of PhysicalDevice in data")
	}

	// construct pairs of start and stop points in the string marking the
	// beginning and end of a single PhysicalDevice entry
	for idx, devIdx := range devStart {
		devStartIdx = append(devStartIdx, devIdx[0])

		if idx > 0 {
			// 0, 1
			devEndIdx = append(devEndIdx, devIdx[0]-1)
		}
	}
	devEndIdx = append(devEndIdx, len(output))

	deviceRaw := []string{}
	for idx, devStart := range devStartIdx {
		devEnd := devEndIdx[idx]
		deviceRaw = append(deviceRaw, strings.TrimSpace(output[devStart:devEnd]))
	}

	for _, deviceStr := range deviceRaw {
		deviceLines := strings.SplitN(deviceStr, "\n", 2)
		if len(deviceLines) < 2 {
			return []PhysicalDevice{}, fmt.Errorf("expected more than 2 lines of data, found %d", len(deviceRaw))
		}

		toks := strings.Split(deviceLines[0], "#")
		pdID, err := strconv.Atoi(toks[1])

		if err != nil {
			return []PhysicalDevice{}, fmt.Errorf("error parsing PhysicalDevice device id in %q: %s", toks[1], err)
		}

		pd := PhysicalDevice{
			ID: pdID,
		}

		for _, lineRaw := range strings.Split(deviceLines[1], "\n") {
			line := strings.TrimSpace(lineRaw)

			// ignore lines that didn't have colon in it or have > 1 colon
			rawToks := strings.SplitN(line, ":", 2)
			if len(rawToks) < 2 || len(rawToks) > 2 {
				continue
			}

			// the raw tokens split on colon will have whitespace, so let's trim
			// that into our final key, value tokens:
			// 'Logical Device Name               ',  'LogicalDrv 0'  =>  'Logical Device Name', 'LogicalDrv 0'
			toks := []string{}
			for _, tok := range rawToks {
				toks = append(toks, strings.TrimSpace(tok))
			}

			// skip tokens if key is not in the parse list
			if _, ok := physicalDeviceParseKeys[toks[0]]; !ok {
				continue
			}

			switch {
			case toks[0] == "Block Size":
				dToks := strings.Split(toks[1], " ")

				bSize, err := strconv.Atoi(dToks[0])
				if err != nil {
					return []PhysicalDevice{}, fmt.Errorf("failed to parse Block Size from token %q: %s", dToks[0], err)
				}

				pd.BlockSize = bSize
			case toks[0] == "Physical Block Size":
				dToks := strings.Split(toks[1], " ")

				bSize, err := strconv.Atoi(dToks[0])
				if err != nil {
					return []PhysicalDevice{}, fmt.Errorf("failed to parse Physical Block Size from token %q: %s", dToks[0], err)
				}

				pd.PhysicalBlockSize = bSize
			case toks[0] == "Array":
				aID, err := strconv.Atoi(toks[1])
				if err != nil {
					return []PhysicalDevice{}, fmt.Errorf("failed to parse Array from token %q: %s", toks[1], err)
				}

				pd.ArrayID = aID
			case toks[0] == "Vendor":
				pd.Vendor = toks[1]
			case toks[0] == "Model":
				pd.Model = toks[1]
			case toks[0] == "Firmware":
				pd.Firmware = toks[1]
			case toks[0] == "Serial number":
				pd.SerialNumber = toks[1]
			case toks[0] == "SSD":
				if toks[1] == "Yes" {
					pd.Type = SSD
				} else {
					pd.Type = HDD
				}
			case toks[0] == "State":
				pd.Availability = toks[1]
			case toks[0] == "Total Size":
				dToks := strings.Split(toks[1], " ")
				bSize, err := strconv.Atoi(dToks[0])

				if err != nil {
					return []PhysicalDevice{}, fmt.Errorf("failed to parse Total Size from token %q: %s", dToks[0], err)
				}

				pd.SizeMB = bSize
			case toks[0] == "Write Cache":
				pd.WriteCache = toks[1]
			}
		}

		// ignore any Device that doesn't have a Availability/State
		if len(pd.Availability) > 0 {
			pDevs = append(pDevs, pd)
		}
	}

	return pDevs, nil
}

func parseGetConf(output string) ([]LogicalDevice, []PhysicalDevice, error) {
	logDevs := []LogicalDevice{}
	phyDevs := []PhysicalDevice{}

	if len(output) < 1 {
		return logDevs, phyDevs, fmt.Errorf("cannot parse an empty string")
	}

	lines := strings.Split(output, "\n\n\n")
	if len(lines) < 3 {
		return logDevs, phyDevs, fmt.Errorf("expected more than 3 lines of data in input")
	}

	for _, device := range lines {
		ldStart := strings.Index(device, "Logical Device number")
		if ldStart >= 0 {
			ldRawData := strings.TrimSpace(device[ldStart:])

			// parse singular logical device
			logDev, err := parseLogicalDeviceString(ldRawData)
			if err != nil {
				return []LogicalDevice{}, []PhysicalDevice{}, fmt.Errorf("error parsing logical device from arcconf getconfig output: %s", err)
			}

			logDevs = append(logDevs, logDev)
		}

		// all logical devices will have been parsed once we find the Physical Device Info section
		pdStart := strings.Index(device, "Physical Device information")
		if pdStart >= 0 {
			pdRawData := strings.TrimSpace(device[pdStart:])

			// parse all physical devices
			pDevs, err := parsePhysicalDevices(pdRawData)
			if err != nil {
				return []LogicalDevice{}, []PhysicalDevice{}, fmt.Errorf("error parsing physical device from arcconf getconfig output: %s", err)
			}

			phyDevs = pDevs
		}
	}

	return logDevs, phyDevs, nil
}

// parseList - parse arcconf list command output
func parseList(output string) ([]int, error) {
	var numCtrlRe = regexp.MustCompile(`(?m)Controllers found:\s\d+`)
	var ctrlRe = regexp.MustCompile(`(?m)Controller\s\d+`)
	var numCtrls int
	var controllers []int

	// extract the number of controllers that arcconf reported
	numCtrlsStr := numCtrlRe.FindString(output)
	if len(numCtrlsStr) == 0 {
		return controllers, fmt.Errorf("error parsing arcconf output, missing 'Controllers found:' output")
	}

	toks := strings.Split(numCtrlsStr, " ")
	if len(toks) != 3 {
		return controllers, fmt.Errorf("error parsing arcconf output, found %d tokens expected 3", len(toks))
	}

	numCtrls, err := strconv.Atoi(toks[2])
	if err != nil {
		return controllers, fmt.Errorf("failed to parse int from %q: %s", toks[2], err)
	}

	// extract the Controller Ids listed
	for _, match := range ctrlRe.FindAllString(output, -1) {
		toks := strings.Split(match, " ")
		if len(toks) == 2 {
			ctrlID, err := strconv.Atoi(toks[1])
			if err != nil {
				return controllers, fmt.Errorf("failed to parse int from %q: %s", toks[1], err)
			}

			controllers = append(controllers, ctrlID)
		}
	}

	if len(controllers) != numCtrls {
		return []int{}, fmt.Errorf("mismatched output, found %d controllers, expected %d", len(controllers), numCtrls)
	}

	return controllers, nil
}

type arcConf struct {
}

// arcConf returns a arcconf specific implementation of SmartPqi interface
func ArcConf() SmartPqi {
	return &arcConf{}
}

// SmartPqi Interface Implementation
func (ac *arcConf) List() ([]int, error) {
	var stdout, stderr []byte
	var rc int
	var controllerIDs []int

	args := []string{"list", "nologs"}
	if stdout, stderr, rc = arcconf(args...); rc != 0 {
		var err error = ErrNoArcconf
		if rc != noArcConfRC {
			err = cmdError(args, stdout, stderr, rc)
		}

		return controllerIDs, err
	}

	controllerIDs, err := parseList(string(stdout))
	if err != nil {
		return controllerIDs, fmt.Errorf("failed to parse arcconf output: %s", err)
	}

	return controllerIDs, nil
}

func (ac *arcConf) Query(cID int) (Controller, error) {
	ctrlIDs, err := ac.List()
	if err != nil {
		return Controller{}, fmt.Errorf("failed to enumerate controllers: %s", err)
	}

	for _, ctrlID := range ctrlIDs {
		if ctrlID == cID {
			return ac.GetConfig(cID)
		}
	}

	return Controller{}, fmt.Errorf("unknown controller id %d", cID)
}

func (ac *arcConf) GetDiskType(path string) (disko.DiskType, error) {
	cIDs, err := ac.List()
	if err != nil {
		return disko.HDD, fmt.Errorf("failed to enumerate controllers: %s", err)
	}

	errors := []error{}
	for _, cID := range cIDs {
		ctrl, err := ac.GetConfig(cID)
		if err != nil {
			errors = append(errors, fmt.Errorf("error while getting config for controller id:%d: %s", cID, err))
		}
		for _, lDrive := range ctrl.LogicalDrives {
			if lDrive.DiskName == path && lDrive.IsSSD() {
				return disko.SSD, nil
			}

			return disko.HDD, nil
		}
	}

	for _, err := range errors {
		if err != ErrNoArcconf && err != ErrNoController && err != ErrUnsupported {
			return disko.HDD, err
		}
	}

	return disko.HDD, fmt.Errorf("cannot determine disk type")
}

func (ac *arcConf) DriverSysfsPath() string {
	return SysfsPCIDriversPath
}

// not implemented at the driver level
func (ac *arcConf) IsSysPathRAID(path string) bool {
	return false
}

func (ac *arcConf) GetConfig(cID int) (Controller, error) {
	var stdout, stderr []byte
	var rc int

	// getconfig ID
	args := []string{"getconfig", fmt.Sprintf("%d", cID), "nologs"}
	if stdout, stderr, rc = arcconf(args...); rc != 0 {
		var err error = ErrNoArcconf
		if rc != noArcConfRC {
			err = cmdError(args, stdout, stderr, rc)
		}

		return Controller{}, err
	}

	getConfigOut := string(stdout)

	return newController(cID, getConfigOut)
}

func newController(cID int, arcGetConfigOut string) (Controller, error) {
	ctrl := Controller{
		ID: cID,
	}

	lDevs, pDevs, err := parseGetConf(arcGetConfigOut)
	if err != nil {
		return Controller{}, fmt.Errorf("failed to parse arcconf getconfig output: %s", err)
	}

	// PD.ID -> PhysicalDevice
	// type DriveSet map[int]*PhysicalDevice
	ctrl.PhysicalDrives = DriveSet{}
	ctrl.LogicalDrives = LogicalDriveSet{}

	for idx := range pDevs {
		pDev := pDevs[idx]
		ctrl.PhysicalDrives[pDev.ID] = &pDev
	}

	for lIdx := range lDevs {
		lDev := lDevs[lIdx]
		ctrl.LogicalDrives[lDev.ID] = &lDev
		for idx := range pDevs {
			pDev := pDevs[idx]
			if lDev.ArrayID == pDev.ArrayID {
				lDev.Devices = append(lDev.Devices, &pDev)
			}
		}
	}

	return ctrl, nil
}

func arcconf(args ...string) ([]byte, []byte, int) {
	cmd := exec.Command("arcconf", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return stdout.Bytes(), stderr.Bytes(), getCommandErrorRCDefault(err, noArcConfRC)
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
