package disko

// DiskFilter is filter function that returns true if the mathing disk is
// accepted false otherwise.
type DiskFilter func(Disk) bool

// VGFilter is filter function that returns true if the mathing vg is
// accepted false otherwise.
type VGFilter func(VG) bool

// PVFilter is filter function that returns true if the mathing pv is
// accepted false otherwise.
type PVFilter func(PV) bool

// System interface provides system level disk and lvm methods that are
// implemented by the specific system.
type System interface {
	// ScanAllDisks scans the system for all available disks and returns a
	// set of disks that are accepted by the filter function. Use this function
	// if you dont know the device paths for the specific disks to be scanned.
	ScanAllDisks(filter DiskFilter) (DiskSet, error)

	// ScanDisks scans the system for disks identified by the specified paths
	// and returns a set of disks that are accepted by the filter function.
	ScanDisks(filter DiskFilter, paths ...string) (DiskSet, error)

	// ScanDisk scans the system for a single disk specified by the device path.
	ScanDisk(path string) (Disk, error)

	// ScanPVs scans the system for all the PVs and returns the set of PVs that
	// are accepted by the filter function.
	ScanPVs(filter PVFilter) (PVSet, error)

	// ScanVGs scans the systems for all the VGs and returns the set of VGs that
	// are accepted by the filter function.
	ScanVGs(filter VGFilter) (VGSet, error)

	// CreatePV creates a PV with specified name.
	CreatePV(name string) (PV, error)

	// CreateVG creates a VG with specified name and adds the provided pvs to
	// this vg.
	CreateVG(name string, pvs ...PV) (VG, error)
}
