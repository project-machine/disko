package disko

// DiskType enumerates supported disk types.
type DiskType int

const (
	// HDD - hard disk drive
	HDD DiskType = iota

	// SSD - solid state disk
	SSD

	// NVME - Non-volatile memory express
	NVME
)

// AttachmentType enumerates the type of device to which the disks are
// attached to in the system.
type AttachmentType int

const (
	// RAID - indicates that the device is attached to RAID card
	RAID AttachmentType = iota

	// ATA - indicates that the device is attached to ATA card
	ATA

	// PCIE - indicates that the device is attached to PCIE card
	PCIE

	// USB - indicates that the device is attached to USB bus
	USB

	// VIRTIO - indicates that the device is attached to virtio.
	VIRTIO

	// IDE - indicates that the device is attached to IDE.
	IDE
)

// DiskSet is a map of the kernel device name and the disk.
type DiskSet map[string]Disk

// Details prints the details of the disks in the disk set ina a tabular
// format.
func (ds DiskSet) Details() string {
	return ""
}

// Disk interface wraps the disk level operations. It provides basic information
// about the disk including name, device path, size etc. Operations include
// creation and deletion of partitions and wiping the disk clean.
type Disk interface {
	// Name returns the kernel name of the disk.
	Name() string

	// Path returns the device path of the disk.
	Path() string

	// Size returns the size of the disk in bytes.
	Size() uint64

	// SectorSize return the sector size of the device, if its unknown or not
	// applicable it will return 0.
	SectorSize() uint

	// FreeSpace returns the slots of free spaces on the disk. These slots can
	// be used to create new partitions.
	FreeSpace() []FreeSpace

	// Type returns the DiskType indicating the type of this disk. This method
	// can be used to determine if the disk is of a particular media type like
	// HDD, SSD or NVMe.
	Type() DiskType

	// Attachement returns the type of storage card this disk is attached to.
	// For example: RAID, ATA or PCIE.
	Attachment() AttachmentType

	// Partitions returns the set of partitions on this disk.
	Partitions() PartitionSet

	// UdevInfo returns the disk's udev information.
	UdevInfo() UdevInfo

	// CreatePartition creates a partition on the is disk with the specified
	// partition number, type and disk offsets.
	CreatePartition(Partition) error

	// DeletePartition deletes the specified partition.
	DeletePartition(int) error

	// Wipe wipes the disk to make it a clean disk. All partitions and data
	// on the disk will be lost.
	Wipe() error
}

// PartitionSet is a map of partition number to the partition.
type PartitionSet map[int]Partition

// Partition interface wraps the disk partition information.
type Partition interface {
	// Start returns the start offset of the disk partition.
	Start() uint64

	// End returns the end offset of the disk partition.
	End() uint64

	// Id returns the partition id.
	Id() string

	// Type returns the partition type.
	Type() string

	// Name returns the name of this partition.
	Name() string

	// Number returns the number of this partition.
	Number() uint

	// Size returns the size of this partition.
	Size() uint64
}

// UdevInfo captures the udev information about a disk.
type UdevInfo struct {
	// Name of the disk
	Name string

	// SysPath is the system path of this device.
	SysPath string

	// Symlinks for the disk.
	Symlinks []string

	// Properties is udev information as a map of key, value pairs.
	Properties map[string]string
}

// FreeSpace indicates a free slot on the disk with a Start and End offset,
// where a partition can be craeted.
type FreeSpace struct {
	Start uint64
	End   uint64
}

// Size returns the size of the free space, which is End - Start.
func (f *FreeSpace) Size() uint64 {
	return f.End - f.Start
}

// ScanSystem scans the system for all the disks and returns a disk set with
// the disk information of all the discovered disks.
func ScanSystem() (DiskSet, error) {
	return nil, nil
}

// ScanDisk will scan for the disk specified by the disk path. Returns the disk
// information if found else error.
func ScanDisk(devicePath string) (Disk, error) {
	return nil, nil
}

// DiskSetByType is a convinient mapping of all the disks by their types.
type DiskSetByType map[DiskType]DiskSet

// MakeDiskSetByType take a set of disks, groups them by type and returns
// the group. This is useful helper routing to scall all disks in the system
// and group them by their type.
func MakeDiskSetByType(allDisks DiskSet) DiskSetByType {
	disksByType := make(DiskSetByType)
	for n, d := range allDisks {
		if tset, ok := disksByType[d.Type()]; ok {
			tset[n] = d
		} else {
			tset := make(DiskSet)
			tset[n] = d
			disksByType[d.Type()] = tset
		}
	}
	return disksByType
}

// ScanDiskByType scans the system for the specified disk type and returns the
// disks.
func ScanDiskByType(diskType DiskType) (DiskSet, error) {
	return nil, nil
}
