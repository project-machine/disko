package mpi3mr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"machinerun.io/disko"
)

// parse JSON from 'storcli2 show nolog J' for List() method
type StorCli2CmdShowResult struct {
	Controllers []StorCli2ShowCtrl `json:"Controllers"`
}

type StorCli2ShowCtrl struct {
	CommandStatus CommandStatus    `json:"Command Status"`
	ResponseData  ResponseDataShow `json:"Response Data"`
}

type ResponseDataShow struct {
	NumberOfControllers int              `json:"Number of Controllers,omitempty"`
	HostName            string           `json:"Host Name,omitempty"`
	OperatingSystem     string           `json:"Operating System,omitempty"`
	SL8LibraryVersion   string           `json:"SL8 Library Version,omitempty"`
	SystemOverview      []SystemOverview `json:"System Overview,omitempty"`
}

type SystemOverview struct {
	Ctrl         int    `json:"Ctrl,omitempty"`
	ProductName  string `json:"Product Name,omitempty"`
	SASAddress   string `json:"SASAddress,omitempty"`
	Status       string `json:"Status,omitempty"`
	PDs          int    `json:"PD(s),omitempty"`
	VDs          int    `json:"VD(s),omitempty"`
	VNOpt        int    `json:"VNOpt,omitempty"`
	EPack        string `json:"EPack,omitempty"`
	SerialNumber string `json:"SerialNumber,omitempty"`
}

// parse JSON from 'storcli2 /cX show nolog J' for Query() method
type StorCli2CmdCxShowResult struct {
	Controllers []StorCli2CxShowCtrl `json:"Controllers"`
}

type StorCli2CxShowCtrl struct {
	CommandStatus CommandStatus      `json:"Command Status"`
	ResponseData  ResponseDataCxShow `json:"Response Data"`
}

type ResponseDataCxShow struct {
	DriveGroupCount    int             `json:"Drive Groups,omitempty"`
	DriveGroups        []DriveGroup    `json:"TOPOLOGY,omitempty"`
	VirtualDriveCount  int             `json:"Virtual Drives,omitempty"`
	VirtualDrives      []VirtualDrive  `json:"VD LIST,omitempty"`
	PhysicalDriveCount int             `json:"Physical Drives"`
	PhysicalDrives     []PhysicalDrive `json:"PD LIST,omitempty"`
}

// parse JSON from 'storcli2 /cX/vall show all nolog J' for Query() method
type StorCli2CmdCxVallShowAllResult struct {
	Controllers []StorCli2CxVallShowAllCtrl `json:"Controllers"`
}

type StorCli2CxVallShowAllCtrl struct {
	CommandStatus CommandStatus             `json:"Command Status"`
	ResponseData  ResponseDataCxVallShowAll `json:"Response Data"`
}

type ResponseDataCxVallShowAll struct {
	VirtualDrives []VirtualDrives `json:"Virtual Drives,omitempty"`
	DriveGroups   []DriveGroup    `json:"TOPOLOGY,omitempty"`
	VDList        []VirtualDrive  `json:"VD LIST,omitempty"`
	VDCount       int             `json:"Total VD Count,omitempty"`
	PDList        []PhysicalDrive `json:"PD LIST,omitempty"`
	PDCount       int             `json:"Physical Drives"`
	DGDriveList   []PhysicalDrive `json:"DG Drive List,omitempty"`
	DriveCount    int             `json:"Total Drive Count,omitempty"`
}

// common in all commands
type CommandStatus struct {
	CLIVersion      string `json:"CLI Version"`
	OperatingSystem string `json:"Operating system"`
	Controller      string `json:"Controller"`
	Status          string `json:"Status"`
	Description     string `json:"Description"`
}

type DriveGroup struct {
	DG      int    `json:"DG,omitempty"`
	Array   int    `json:"Arr,omitempty"`
	Row     int    `json:"Row,omitempty"`
	EIDSlot string `json:"EID:Slot,omitempty"`
	PID     int    `json:"PID,omitempty"`
	DID     int    `json:"DID,omitempty"`
	Type    string `json:"Type,omitempty"`
	State   string `json:"State,omitempty"`
	Status  string `json:"Status,omitempty"`
	BT      string `json:"BT,omitempty"`
	Size    string `json:"Size,omitempty"`
	PDC     string `json:"PDC,omitempty"`
	PI      string `json:"PI,omitempty"`
	SED     string `json:"SED,omitempty"`
	DS3     string `json:"DS3,omitempty"`
	Secured string `json:"Secured,omitempty"`
	FSpace  string `json:"FSpace,omitempty"`
}

func (dg *DriveGroup) UnmarshalJSON(data []byte) error {
	tmp := make(map[string]interface{})
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	for k, v := range tmp {
		switch k {
		case "DG":
			if v == "-" {
				dg.DG = int(-1)
			} else {
				dg.DG = int(v.(float64))
			}
		case "Arr":
			if v == "-" {
				dg.Array = int(-1)
			} else {
				dg.Array = int(v.(float64))
			}
		case "Row":
			if v == "-" {
				dg.Row = int(-1)
			} else {
				dg.Row = int(v.(float64))
			}
		case "EID:Slot":
			dg.EIDSlot = v.(string)
		case "DID":
			if v == "-" {
				dg.DID = int(-1)
			} else {
				dg.DID = int(v.(float64))
			}
		case "Type":
			dg.Type = v.(string)
		case "State":
			dg.State = v.(string)
		case "BT":
			dg.BT = v.(string)
		case "Size":
			dg.Size = v.(string)
		case "PDC":
			dg.PDC = v.(string)
		case "PI":
			dg.PI = v.(string)
		case "SED":
			dg.SED = v.(string)
		case "DS3":
			dg.DS3 = v.(string)
		case "FSpace":
			dg.FSpace = v.(string)
		default:
		}
	}
	return nil
}

type VirtualDrives struct {
	VDInfo         VirtualDrive           `json:"VD Info"`
	PhysicalDrives []PhysicalDrive        `json:"PDs"`
	VDProperties   VirtualDriveProperties `json:"VD Properties"`
}

/*
VD Info

	"VD Info": {
	  "DG/VD": "0/1",
	  "TYPE": "RAID0",
	  "State": "Optl",
	  "Access": "RW",
	  "CurrentCache": "NR,WB",
	  "DefaultCache": "NR,WB",
	  "Size": "893.137 GiB",
	  "Name": ""
	}
*/
type VirtualDrive struct {
	DGVD           string                 `json:"DG/VD,omitempty"`
	Type           string                 `json:"TYPE,omitempty"`
	State          string                 `json:"State,omitempty"`
	Access         string                 `json:"Access,omitempty"`
	CurrentCache   string                 `json:"CurrentCache,omitempty"`
	DefaultCache   string                 `json:"DefaultCache,omitempty"`
	Size           string                 `json:"Size,omitempty"`
	Name           string                 `json:"Name,omitempty"`
	Properties     VirtualDriveProperties `json:"Properties,omitempty"`
	PhysicalDrives []PhysicalDrive        `json:"PhysicalDrives,omitempty"`
}

func (vd *VirtualDrive) ID() string {
	return strings.Replace(vd.DGVD, "/", "-", -1) // 0/1 -> 0-1
}

func (vd *VirtualDrive) DG() (int, error) {
	dgvd := strings.Split(vd.DGVD, "/")
	if len(dgvd) != 2 {
		return -1, errors.Errorf("failed to parse DriveGroup ID from %q", vd.DGVD)
	}
	return parseIntOrDash(dgvd[0])
}

func (vd *VirtualDrive) VD() (int, error) {
	dgvd := strings.Split(vd.DGVD, "/")
	if len(dgvd) != 2 {
		return -1, errors.Errorf("failed to parse VirtualDrive ID from %q", vd.DGVD)
	}
	return parseIntOrDash(dgvd[1])
}

func (vd *VirtualDrive) Path() string {
	return vd.Properties.OSDriveName
}

/*
Physical Drive JSON

	{
		"EID:Slt": "322:2",
		"PID": 312,
		"State": "Conf",
		"Status": "Online",
		"DG": 0,
		"Size": "893.137 GiB",
		"Intf": "SATA",
		"Med": "SSD",
		"SED_Type": "-",
		"SeSz": "512B",
		"Model": "SAMSUNG MZ7L3960HCJR-00AK1",
		"Sp": "U",
		"LU/NS Count": 1,
		"Alt-EID": "-"
	}
*/
type PhysicalDrive struct {
	EIDSlot    string `json:"EID:Slt"`
	PID        int    `json:"PID"`
	State      string `json:"State"`
	Status     string `json:"Status"`
	DG         int    `json:"DG"`
	Size       string `json:"Size"`
	Interface  string `json:"Intf"`
	Medium     string `json:"Med"`
	SEDType    string `json:"SED_Type"`
	SectorSize string `json:"SeSz"`
	Model      string `json:"Model"`
	SP         string `json:"Sp"`
	LUNSCount  int    `json:"LU/NS Count"`
	AltEID     string `json:"Alt-EID"`
}

func (pd *PhysicalDrive) ID() int {
	return pd.PID
}

func (pd *PhysicalDrive) UnmarshalJSON(data []byte) error {
	tmp := make(map[string]interface{})
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	for k, v := range tmp {
		switch k {
		case "EID:Slt":
			pd.EIDSlot = v.(string)
		case "PID":
			//nolint: gosimple
			switch v.(type) {
			case int:
				pd.PID = int(v.(int))
			case string:
				if v == "-" {
					pd.PID = -1
				} else {
					i, err := strconv.Atoi(string(v.(string)))
					if err != nil {
						return err
					}
					pd.PID = i
				}
			}
		case "State":
			pd.State = v.(string)
		case "Status":
			pd.Status = v.(string)
		case "DG":
			//  S1034: assigning the result of this type assertion to a variable
			//  (switch v := v.(ype)) could eliminate type assertions in switch
			//  cases (gosimple)  ? wtf
			//nolint: gosimple
			switch v.(type) {
			case int:
				pd.DG = v.(int)
			case string:
				if v == "-" {
					pd.DG = -1
				} else {
					i, err := strconv.Atoi(v.(string))
					if err != nil {
						return err
					}
					pd.DG = i
				}
			}
		case "Size":
			pd.Size = v.(string)
		case "Intf":
			pd.Interface = v.(string)
		case "Med":
			pd.Medium = v.(string)
		case "SED_Type":
			pd.SEDType = v.(string)
		case "SeSz":
			pd.SectorSize = v.(string)
		case "Model":
			pd.Model = v.(string)
		case "Sp":
			pd.SP = v.(string)
		case "LU/NS Count":
			pd.LUNSCount = int(v.(float64))
		case "Alt-EID":
			pd.AltEID = v.(string)
		}
	}
	return nil
}

func (pd *PhysicalDrive) EID() (int, error) {
	toks := strings.SplitN(pd.EIDSlot, ":", 2)
	if len(toks) != 2 {
		return -1, errors.Errorf("failed to parse EID ID from %q", pd.EIDSlot)
	}

	return parseIntOrDash(toks[0])
}

func (pd *PhysicalDrive) Slot() (int, error) {
	toks := strings.SplitN(pd.EIDSlot, ":", 2)
	if len(toks) != 2 {
		return -1, errors.Errorf("failed to parse Slot ID from %q", pd.EIDSlot)
	}

	return parseIntOrDash(toks[1])
}

type VirtualDriveProperties struct {
	StripSize                    string `json:"Strip Size,omitempty"`
	BlockSize                    int    `json:"Block Size,omitempty"`
	NumberOfBlocks               int    `json:"Number of Blocks,omitempty"`
	SpanDepth                    int    `json:"Span Depth,omitempty"`
	NumberOfDrives               int    `json:"Number of Drives,omitempty"`
	DriveWriteCachePolicy        string `json:"Drive Write Cache Policy,omitempty"`
	DefaultPowerSavePolicy       string `json:"Default Power Save Policy,omitempty"`
	CurrentPowerSavePolicy       string `json:"Current Power Save Policy,omitempty"`
	AccessPolicyStatus           string `json:"Access Policy Status,omitempty"`
	AutoBGI                      string `json:"Auto BGI,omitempty"`
	Secured                      string `json:"Secured,omitempty"`
	InitState                    string `json:"Init State,omitempty"`
	Consistent                   string `json:"Consistent,omitempty"`
	Morphing                     string `json:"Morphing,omitempty"`
	CachePreserved               string `json:"Cache Preserved,omitempty"`
	BadBlockExists               string `json:"Bad Block Exists,omitempty"`
	VDReadyForOSRequests         string `json:"VD Ready for OS Requests,omitempty"`
	ReachedLDBBMFailureThreshold string `json:"Reached LD BBM failure threshold,omitempty"`
	SupportedEraseTypes          string `json:"Supported Erase Types,omitempty"`
	ExposedToOS                  string `json:"Exposed to OS,omitempty"`
	CreationTimeLocalTimeString  string `json:"Creation Time(LocalTime yyyy/mm/dd hh:mm:sec),omitempty"`
	DefaultCachebypassMode       string `json:"Default Cachebypass Mode,omitempty"`
	CurrentCachebypassMode       string `json:"Current Cachebypass Mode,omitempty"`
	SCSINAAId                    string `json:"SCSI NAA Id,omitempty"`
	OSDriveName                  string `json:"OS Drive Name,omitempty"`
	CurrentUnmapStatus           string `json:"Current Unmap Status,omitempty"`
	CurrentWriteSameUnmapStatus  string `json:"Current WriteSame Unmap Status,omitempty"`
	LUNSCountUsedPerPD           int    `json:"LU/NS Count used per PD,omitempty"`
	DataFormatForIO              string `json:"Data Format for I/O,omitempty"`
	SerialNumber                 string `json:"Serial Number,omitempty"`
}

func newController(cID int, cxShowNoLogJOut []byte, cxVAllShowAllJOut []byte) (Controller, error) {
	ctrl := Controller{
		ID: cID,
	}

	// parse Physical and Virtual Drives
	s := StorCli2CmdCxShowResult{}
	if err := json.Unmarshal(cxShowNoLogJOut, &s); err != nil {
		return ctrl, err
	}

	// parse VirtDriveProperties
	v := StorCli2CmdCxVallShowAllResult{}
	if err := json.Unmarshal(cxVAllShowAllJOut, &v); err != nil {
		return ctrl, err
	}

	// collect the physical drives
	physDrives := []PhysicalDrive{}
	foundPDrives := s.Controllers[cID].ResponseData.PhysicalDrives
	physDrives = append(physDrives, foundPDrives...)

	// collect the virtual drives
	virtDrives := []VirtualDrive{}
	foundVDrives := v.Controllers[cID].ResponseData.VirtualDrives // type VirtualDrives{}

	// collect the virtual drives from the c0/vall showall output, and insert
	// the properties into the private field of the VirtualDrive
	for vID := range foundVDrives {
		vdrivesInfo := foundVDrives[vID]

		vdrive := vdrivesInfo.VDInfo // this is a struct VirtualDrive
		vdrive.Properties = vdrivesInfo.VDProperties
		vdrive.PhysicalDrives = vdrivesInfo.PhysicalDrives

		virtDrives = append(virtDrives, vdrive)
	}

	// populate the controller
	ctrl.PhysicalDrives = PhysicalDriveSet{}
	ctrl.VirtualDrives = VirtualDriveSet{}

	// identify physical drives by PID
	for idx := range physDrives {
		pDev := physDrives[idx]
		pid := strconv.Itoa(int(pDev.PID))
		ctrl.PhysicalDrives[pid] = pDev
	}

	// VirtualDrives are indexed by DG/VD string DriveGroupID/VirtualDriveID
	for vIdx := range virtDrives {
		vDev := virtDrives[vIdx]
		ctrl.VirtualDrives[vDev.DGVD] = vDev
	}

	return ctrl, nil
}

func parseIntOrDash(field string) (int, error) {
	if field == "-" {
		return -1, nil
	}

	return strconv.Atoi(field)
}

// parse the controllers listed
// cmdOut is output of `storcli2 show nolog J`
func parseShowOutput(showOutput []byte) ([]int, error) {
	s := StorCli2CmdShowResult{}
	if err := json.Unmarshal(showOutput, &s); err != nil {
		return []int{}, err
	}

	cIDs := []int{}
	for _, ctrl := range s.Controllers {
		for i := 0; i < ctrl.ResponseData.NumberOfControllers; i++ {
			cIDs = append(cIDs, ctrl.ResponseData.SystemOverview[i].Ctrl)
		}
	}

	return cIDs, nil
}

func storcli2(args ...string) ([]byte, []byte, int) {
	cmd := exec.Command("storcli2", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return stdout.Bytes(), stderr.Bytes(), getCommandErrorRCDefault(err, noStorCli2RC)
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

	exitRC := exitError.ExitCode()
	if exitRC == StorCli2ShowRC {
		return 0
	}
	if ok {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return rcError
}

// Implement the mpi3mr Interface with storcli2
type storCli2 struct {
}

// StorCli returns a storcli2 specific implementation of Query for the Mpi3mr interface
func StorCli2() Mpi3mr {
	return &storCli2{}
}

const noStorCli2RC = 127

// when successful storcli2 show returns rc=6  ¯\_(ツ)_/¯
const StorCli2ShowRC = 6

func (sc *storCli2) Query(cID int) (Controller, error) {
	// run /c0 show all nolog J
	//   - get PDs and VDs in JSON format
	// run /c0/vall show all nolog J
	//   - populate VD Properties and Path
	var stdout, stderr []byte
	var rc int

	args := []string{fmt.Sprintf("/c%d", cID), "show", "nolog", "J"}

	if stdout, stderr, rc = storcli2(args...); rc != 0 {
		var err error = ErrNoStor2cli
		if rc != noStorCli2RC {
			err = cmdError(args, stdout, stderr, rc)
		}

		return Controller{}, err
	}

	cxShowNoLogJOut := stdout

	args = []string{fmt.Sprintf("/c%d/vall", cID), "show", "all", "nolog", "J"}
	if stdout, stderr, rc = storcli2(args...); rc != 0 {
		return Controller{}, cmdError(args, stdout, stderr, rc)
	}

	cxVAllShowAllJOut := stdout

	return newController(cID, cxShowNoLogJOut, cxVAllShowAllJOut)
}

func (sc *storCli2) List() ([]int, error) {
	var stdout, stderr []byte
	var rc int

	args := []string{"show", "nolog", "J"}

	if stdout, stderr, rc = storcli2(args...); rc != 0 {
		var err error = ErrNoStor2cli
		if rc != noStorCli2RC {
			err = cmdError(args, stdout, stderr, rc)
		}

		return []int{}, err
	}

	ctrlIDs, err := parseShowOutput(stdout)
	if err != nil {
		return []int{}, errors.Errorf("failed to parse output from 'storcli2 show nolog J: %s", err)
	}

	return ctrlIDs, nil
}

func (sc *storCli2) DriverSysfsPath() string {
	return SysfsPCIDriversPath
}

func (sc *storCli2) GetDiskType(path string) (disko.DiskType, error) {
	cIDs, err := sc.List()
	if err != nil {
		return disko.HDD, errors.Errorf("failed to get controller list: %s", err)
	}

	errors := []error{}
	for _, cID := range cIDs {
		ctrl, err := sc.Query(cID)
		if err != nil {
			errors = append(errors, fmt.Errorf("error while getting config for controller id:%d %s", cID, err))
			continue
		}
		for _, vDev := range ctrl.VirtualDrives {
			if vDev.Path() == path && vDev.IsSSD() {
				return disko.SSD, nil
			}

			return disko.HDD, nil
		}
	}

	for _, err := range errors {
		if err != ErrNoStor2cli && err != ErrNoController && err != ErrUnsupported {
			return disko.HDD, err
		}
	}

	return disko.HDD, fmt.Errorf("cannot determine diskt type for path %q", path)
}

// not implemented in driver layer
func (sc *storCli2) IsSysPathRAID(syspath string) bool {
	return false
}
