package disko

import (
	"encoding/json"
	"fmt"
)

// VolumeManager provides logical volume oprations that allows for creation and
// management of volume groups, physical volumes and logical volumes.
type VolumeManager interface {
	// ScanPVs scans the system for all the PVs and returns the set of PVs that
	// are accepted by the filter function.
	ScanPVs(filter PVFilter) (PVSet, error)

	// ScanVGs scans the systems for all the VGs and returns the set of VGs that
	// are accepted by the filter function.
	ScanVGs(filter VGFilter) (VGSet, error)

	// CreatePV creates a PV with specified name.
	CreatePV(diskName string) (PV, error)

	// DeletePV deletes the specified PV.
	DeletePV(pv PV) error

	// HasPV returns true if the pv exists. This indicates that the device
	// already has an lvm pv header.
	HasPV(name string) bool

	// CreateVG creates a VG with specified name and adds the provided pvs to
	// this vg.
	CreateVG(name string, pvs ...PV) (VG, error)

	// ExtendVG extends the volument group storage capacity with the specified
	// PVs.
	ExtendVG(vgName string, pvs ...PV) error

	// Delete deletes this VG and all the LVs in the VG.
	RemoveVG(vgName string) error

	// HasVG returns true if the vg exists.
	HasVG(vgName string) bool

	// CryptFormat setups up encryption for this volume using the provided key.
	CryptFormat(vgName string, lvName string, key string) error

	// CryptOpen opens the encrypted logical volume for use using the provided
	// key.
	CryptOpen(vgName string, lvName string, decryptedName string, key string) error

	// CryptClose close the encrypted logical volume using the provided key.
	CryptClose(vgName string, lvName string, decryptedName string) error

	// CreateLV creates a LV with specified name, size and type.
	CreateLV(vgName string, name string, size uint64, lvType LVType) (LV, error)

	// RemoveLV removes this LV.
	RemoveLV(vgName string, lvName string) error

	// RenameLV renames this LV to newLvName.
	RenameLV(vgName string, lvName string, newLvName string) error

	// ExtendLV expands the LV to the requested new size.
	ExtendLV(vgName string, lvName string, newSize uint64) error

	// HasVG returns true if the lv exists.
	HasLV(vgName string, name string) bool
}

// PV wraps a LVM physical volume. A lvm physical volume is the raw
// block device or other disk like devices that provide storage capacity.
type PV struct {
	// Name returns the name of the PV.
	Name string `json:"name"`

	// UUID for the PV
	UUID string `json:"uuid"`

	// Path returns the device path of the PV.
	Path string `json:"path"`

	// Size returns the size of the PV.
	Size uint64 `json:"size"`

	// The volume group this PV is part of ("" if none)
	VGName string `json:"vgname"`

	// FreeSize returns the free size of the PV.
	FreeSize uint64 `json:"freeSize"`
}

// PVSet is a set of PVs indexed by their names.
type PVSet map[string]PV

// LVSet is a map of LV names to the LV.
type LVSet map[string]LV

// ExtentSize is extent size for lvm
const ExtentSize = 4 * Mebibyte

// LV interface wraps the lvm logical volume information and operations. A
// logical volume partitions a volume group into a slice of capacity that can
// be used a block device to create a file system.
type LV struct {
	// Name is the name of the logical volume.
	Name string `json:"name"`

	// UUID for the LV
	UUID string `json:"uuid"`

	// Path is the full path of the logical volume.
	Path string `json:"path"`

	// Size the size of the logical volume.
	Size uint64 `json:"size"`

	// Type is the type of logical volume.
	Type LVType `json:"type"`

	// The volume group that this logical volume is part of.
	VGName string `json:"vgname"`

	// Encrypted indicates if the logical volume is encrypted.
	Encrypted bool `json:"encrypted"`

	// DecryptedLVName is the name of the decrypted logical volume as set by
	// the CryptOpen call.
	DecryptedLVName string `json:"decryptedLVName"`

	// DecryptedLVPath is the full path of the decrypted logical volume. This
	// is set only for encrypted volumes, using the CryptFormat.
	DecryptedLVPath string `json:"decryptedLVPath"`
}

// LVType defines the type of the logical volume.
type LVType int

const (
	// THICK indicates thickly provisioned logical volume.
	THICK LVType = iota

	// THIN indicates thinly provisioned logical volume.
	THIN

	// THINPOOL indicates a pool lv for other lvs
	THINPOOL

	// LVTypeUnknown - unknown type
	LVTypeUnknown
)

//nolint:gochecknoglobals
var lvTypes = map[string]LVType{
	"THICK":    THICK,
	"THIN":     THIN,
	"THINPOOL": THINPOOL,
	"UNKNOWN":  LVTypeUnknown,
}

func (t LVType) String() string {
	for k, v := range lvTypes {
		if v == t {
			return k
		}
	}

	return fmt.Sprintf("UNKNOWN-%d", t)
}

func stringToLVType(s string) LVType {
	if val, ok := lvTypes[s]; ok {
		return val
	}

	return LVTypeUnknown
}

// UnmarshalJSON - custom to read as strings or int
func (t *LVType) UnmarshalJSON(b []byte) error {
	var err error
	var asStr string
	var asInt int

	err = json.Unmarshal(b, &asInt)
	if err == nil {
		*t = LVType(asInt)
		return nil
	}

	err = json.Unmarshal(b, &asStr)
	if err != nil {
		return err
	}

	lvtype := stringToLVType(asStr)
	*t = lvtype

	return nil
}

// MarshalJSON - serialize to json
func (t LVType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// VG wraps a LVM volume group. A volume group combines one or more
// physical volumes into storage pools and provides a unified logical device
// with combined storage capacity of the underlying physical volumes.
type VG struct {
	// Name is the name of the volume group.
	Name string `json:"name"`

	// UUID for the VG
	UUID string `json:"uuid"`

	// Size is the current size of the volume group.
	Size uint64 `json:"size"`

	// Volumes is set of all the volumes in this volume group.
	Volumes LVSet `json:"volumes"`

	// FreeSpace is the amount free space left in the volume group.
	FreeSpace uint64 `json:"freeSpace"`

	// PVs is the set of PVs that belongs to this VG.
	PVs PVSet `json:"pvs"`
}

// VGSet is set of volume groups indexed by their name.
type VGSet map[string]VG

// Details returns a formatted string with the information of volume groups.
func (vgs VGSet) Details() string {
	return ""
}
