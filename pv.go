package disko

// PV interface wraps a LVM physical volume. A lvm physical volume is the raw
// block device or other disk like devices that provide storage capacity.
type PV interface {
	// Name returns the name of the PV.
	Name() string

	// Path returns the device path of the PV.
	Path() string

	// Size returns the size of the PV.
	Size() uint64

	// FreeSize returns the free size of the PV.
	FreeSize() uint64

	// Create creates this this PV.
	Create() error

	// Remove removes this PVrn
	Remove() error

	// Exists returns true if the PV already exits, else false.
	Exists() bool
}

// PVSet is a set of PVs indexed by their names.
type PVSet map[string]PV
