package megaraid

import (
	"encoding/json"
	"errors"
)

// Controller - a Megaraid controller
type Controller struct {
	ID          int
	Drives      DriveSet
	DriveGroups DriveGroupSet
	VirtDrives  VirtDriveSet
}

// VirtDrive - represents a virtual drive.
type VirtDrive struct {
	// the Virtual Drive Number / ID
	ID int
	// the Drive Group ID
	DriveGroup int
	// Path in linux - may be empty if "Exposed to OS" != "Yes"
	Path string
	// "Name" in output - exposed in cimc UI
	RaidName string
	// Type RAID type (RAID0, RAID1...)
	Type string
	// /c0/v0 data as a string map
	Raw map[string]string
	// "VD0 Properties" as a map
	Properties map[string]string
}

// VirtDriveSet - a map of VirtDrives by their Number.
type VirtDriveSet map[int]*VirtDrive

// DriveGroup - a megaraid "Drive Group". These really have nothing
// that is their own other than their ID.
type DriveGroup struct {
	ID     int
	Drives DriveSet
}

// IsSSD - is this drive group composed of all SSD
func (dg *DriveGroup) IsSSD() bool {
	if len(dg.Drives) == 0 {
		return false
	}

	for _, drive := range dg.Drives {
		if drive.MediaType != SSD {
			return false
		}
	}

	return true
}

// DriveGroupSet - map of DriveGroups by their ID
type DriveGroupSet map[int]*DriveGroup

// MarshalJSON - serialize to json.  Custom Marshal to only reference Disks
// by ID not by full dump.
func (dgs DriveGroupSet) MarshalJSON() ([]byte, error) {
	type terseDriveGroupSet struct {
		ID     int
		Drives []int
	}

	var mySet = []terseDriveGroupSet{}

	for id, driveSet := range dgs {
		drives := []int{}
		for drive := range driveSet.Drives {
			drives = append(drives, drive)
		}

		mySet = append(mySet, terseDriveGroupSet{ID: id, Drives: drives})
	}

	return json.Marshal(&mySet)
}

// Drive - a megaraid (physical) Drive.
type Drive struct {
	ID         int
	DriveGroup int
	EID        int
	Slot       int
	State      string
	MediaType  MediaType
	Model      string
	Raw        map[string]string
}

// DriveSet - just a map of Drives by ID
type DriveSet map[int]*Drive

// IsEqual - compare the two drives (does not compare Raw)
func (d *Drive) IsEqual(d2 Drive) bool {
	return (d.ID == d2.ID && d.EID == d2.EID && d.Slot == d2.Slot &&
		d.DriveGroup == d2.DriveGroup && d.State == d2.State &&
		d.MediaType == d2.MediaType &&
		d.Model == d2.Model)
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

// MegaRaid - basic interface
type MegaRaid interface {
	// Query - Query the controller provided
	Query(int) (Controller, error)
}

// ErrNoController - Error reported by Query if no controller is found.
var ErrNoController = errors.New("megaraid Controller not found")

// ErrUnsupported - Error reported by Query if controller is not supported.
var ErrUnsupported = errors.New("megaraid Controller unsupported")

// ErrNoStorcli - Error reported by Query if no storcli binary in PATH
var ErrNoStorcli = errors.New("no 'storcli' command in PATH")
