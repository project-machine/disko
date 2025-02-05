package mpi3mr

import (
	"encoding/json"
	"errors"

	"github.com/google/go-cmp/cmp"
	"machinerun.io/disko"
)

// Controller - a mpi3mr controller
type Controller struct {
	ID             int              `json:"ID"`
	PhysicalDrives PhysicalDriveSet `json:"PhysicalDrives"`
	VirtualDrives  VirtualDriveSet  `json:"VirtualDrives"`
}

type PhysicalDriveSet map[string]PhysicalDrive
type VirtualDriveSet map[string]VirtualDrive

// IsSSD - is this drive group composed of all SSD
func (vd *VirtualDrive) IsSSD() bool {
	if len(vd.PhysicalDrives) == 0 {
		return false
	}
	for pID := range vd.PhysicalDrives {
		if vd.PhysicalDrives[pID].Medium != "SSD" {
			return false
		}
	}
	return true
}

// IsEqual - compare the two drives
func (vd *VirtualDrive) IsEqual(vd2 VirtualDrive) bool {
	return cmp.Equal(*vd, vd2)
}

// MediaType - a disk "Media"
type MediaType int

const (
	// UnknownMedia - indicates an unknown media
	UnknownMedia MediaType = iota

	// HDD - Spinning hard disk.
	HDD

	// SSD - Solid State Disk
	SSD
)

func (t MediaType) String() string {
	return []string{"UNKNOWN", "HDD", "SSD"}[t]
}

// MarshalJSON for string output rather than int
func (t MediaType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// Mpi3mr - basic interface
type Mpi3mr interface {
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
var ErrNoController = errors.New("mpi3mr Controller not found")

// ErrUnsupported - Error reported by Query if controller is not supported.
var ErrUnsupported = errors.New("mpi3mr Controller unsupported")

// ErrNoStor2cli - Error reported by Query if no storcli binary in PATH
var ErrNoStor2cli = errors.New("no 'storcli2' command in PATH")
