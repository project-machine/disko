package smartpqi

import (
	"encoding/json"
	"errors"
	"strings"

	"machinerun.io/disko"
)

type Controller struct {
	ID             int
	PhysicalDrives DriveSet
	LogicalDrives  LogicalDriveSet
}

type PhysicalDevice struct {
	ArrayID           int       `json:"ArrayID"`
	Availability      string    `json:"Availability"`
	BlockSize         int       `json:"BlockSize"`
	Channel           int       `json:"Channel"`
	ID                int       `json:"ID"`
	Firmware          string    `json:"Firmware"`
	Model             string    `json:"Model"`
	PhysicalBlockSize int       `json:"PhysicalBlockSize"`
	Protocol          string    `json:"Protocol"`
	SerialNumber      string    `json:"SerialNumber"`
	SizeMB            int       `json:"SizeMB"`
	Type              MediaType `json:"Type"`
	Vendor            string    `json:"Vendor"`
	WriteCache        string    `json:"WriteCache"`
}

type LogicalDevice struct {
	ArrayID       int    `json:"ArrayID"`
	BlockSize     int    `json:"BlockSize"`
	Caching       string `json:"Caching"`
	Devices       []*PhysicalDevice
	DiskName      string `json:"DiskName"`
	ID            int    `json:"ID"`
	InterfaceType string `json:"InterfaceType"`
	Name          string `json:"Name"`
	RAIDLevel     string `json:"RAIDLevel"`
	SizeMB        int    `json:"SizeMB"`
}

// IsSSD - is this logical device composed of all SSD
func (ld *LogicalDevice) IsSSD() bool {
	if len(ld.Devices) == 0 {
		return false
	}

	for _, pDev := range ld.Devices {
		if pDev.Type != SSD {
			return false
		}
	}

	return true
}

type DriveSet map[int]*PhysicalDevice

type LogicalDriveSet map[int]*LogicalDevice

// MediaType
type MediaType int

const (
	// UnknownMedia - indicates an unknown media
	UnknownMedia MediaType = iota

	// HDD - Spinning hard disk.
	HDD

	// SSD - Solid State Disk
	SSD

	// NVME - Non Volatile Memory Express
	NVME
)

func (t MediaType) String() string {
	return []string{"UNKNOWN", "HDD", "SSD", "NVME"}[t]
}

func GetMediaType(mediaType string) MediaType {
	switch strings.ToUpper(mediaType) {
	case "HDD":
		return HDD
	case "SSD":
		return SSD
	case "NVME":
		return NVME
	default:
		return UnknownMedia
	}
}

// MarshalJSON for string output rather than int
func (t MediaType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t *MediaType) UnmarshalJSON(data []byte) error {
	var mt string
	if err := json.Unmarshal(data, &mt); err != nil {
		return err
	}
	*t = GetMediaType(mt)
	return nil
}

// SmartPqi - basic interface
type SmartPqi interface {
	// List  - Return list of Controller IDs
	List() ([]int, error)

	// Query - Query the controller provided
	Query(int) (Controller, error)

	// GetDiskType - Determine the disk type if controller owns disk
	GetDiskType(string) (disko.DiskType, error)

	// DriverSysfsPath - Return the sysfs path to the linux driver for this controller
	DriverSysfsPath() string

	// IsSysPathRAID - Check if sysfs path is a device on the controller
	IsSysPathRAID(string) bool
}

// ErrNoController - Error reported by Query if no controller is found.
var ErrNoController = errors.New("smartpqi Controller not found")

// ErrUnsupported - Error reported by Query if controller is not supported.
var ErrUnsupported = errors.New("smartpqi Controller unsupported")

// ErrNoArcconf - Error reported by Query if no arcconf binary in PATH
var ErrNoArcconf = errors.New("no 'arcconf' command in PATH")
