package disko

// DiskFilter is filter function that returns true if the matching disk is
// accepted false otherwise.
type DiskFilter func(Disk) bool

// VGFilter is filter function that returns true if the matching vg is
// accepted false otherwise.
type VGFilter func(VG) bool

// PVFilter is filter function that returns true if the matching pv is
// accepted false otherwise.
type PVFilter func(PV) bool

// LVFilter is filter function that returns true if the matching lv is
// accepted false otherwise.
type LVFilter func(LV) bool

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

	// CreatePartition creates a partition on the is disk with the specified
	// partition number, type and disk offsets.
	CreatePartition(Disk, Partition) error

	// CreatePartitions creates multiple partitions on disk.
	CreatePartitions(Disk, PartitionSet) error

	// UpdatePartition updates multiple existing partitions on a disk.
	UpdatePartition(Disk, Partition) error

	// UpdatePartitions updates multiple existing partitions on a disk.
	UpdatePartitions(Disk, PartitionSet) error

	// DeletePartition deletes the specified partition.
	DeletePartition(Disk, uint) error

	// Wipe wipes the disk to make it a clean disk. All partitions and data
	// on the disk will be lost.
	Wipe(Disk) error
}
