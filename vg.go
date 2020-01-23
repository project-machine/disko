package disko

// VG interface wraps a LVM volume group. A volume group combines one or more
// physical volumes into storage pools and provides a unified logical device
// with combined storage capacity of the underlying physical volumes.
type VG interface {
	// Name returns the name of the volume group.
	Name() string

	// Size returns the current size of the volume group.
	Size() uint64

	// CreateLV creates a LV with specified name, size and type.
	CreateLV(name string, size uint64, lvType LVType) (LV, error)

	// RemoveLV removes the LV with the specified name in this VG.
	RemoveLV(name string) error

	// Volumes returns all the volumes in this volume group.
	Volumes() LVSet

	// FreeSpace returns the amount free space left in the volume group.
	FreeSpace() uint64

	// Extend extends the volument group storage capacity with the specified
	// PVs.
	Extend(pvs ...PV) error

	// Delete deletes this VG and all the LVs in the VG.
	Delete() error
}

// VGSet is set of volume groups indexed by their name.
type VGSet map[string]VG

// Details returns a formatted string with the information of volume groups.
func (vgs VGSet) Details() string {
	return ""
}
