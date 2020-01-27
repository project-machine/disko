package disko

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
	CryptFormat(lvName string, key string) error

	// CryptOpen opens the encrypted logical volume for use using the provided
	// key.
	CryptOpen(lvName string, key string) error

	// CryptClose close the encrypted logical volume using the provided key.
	CryptClose(lvName string, key string) error

	// CreateLV creates a LV with specified name, size and type.
	CreateLV(vgName string, name string, size uint64, lvType LVType) (LV, error)

	// RemoveLV removes this LV.
	RemoveLV(lvName string) error

	// ExtendLV expands the LV to the requested new size.
	ExtendLV(lvName string, newSize uint64) error

	// HasVG returns true if the lv exists.
	HasLV(name string) bool
}

// PV wraps a LVM physical volume. A lvm physical volume is the raw
// block device or other disk like devices that provide storage capacity.
type PV struct {
	// Name returns the name of the PV.
	Name string `json:"name"`

	// Path returns the device path of the PV.
	Path string `json:"path"`

	// Size returns the size of the PV.
	Size uint64 `json:"size"`

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

	// Size the size of the logical volume.
	Size uint64 `json:"size"`

	// Type is the type of logical volume.
	Type LVType `json:"type"`

	// Encrypted indicates if the logical volume is encrypted.
	Encrypted bool `json:"encrypted"`
}

// LVType defines the type of the logical volume.
type LVType int

const (
	// THICK indicates thickly provisioned logical volume.
	THICK LVType = iota

	// THIN indicates thinly provisioned logical volume.
	THIN
)

// VG wraps a LVM volume group. A volume group combines one or more
// physical volumes into storage pools and provides a unified logical device
// with combined storage capacity of the underlying physical volumes.
type VG struct {
	// Name is the name of the volume group.
	Name string `json:"name"`

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
