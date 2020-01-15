package disko

// DiskType enumerates supported disk types
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
// attached to in the system
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

}

// Disk interface wraps the disk level operations. 
type Disk interface {
	Name() string
	Path() string
	Size() uint64
	SectorSize() uint
	FreeSpace() []FreeSpace
	Type() DiskType
	Attachment() AttachmentType
	Partitions() PartitionSet
	UdevInfo() UdevInfo
	CreatePartition(Partition) error
	DeletePartition(int)
	Wipe() error
}

type PartitionSet map[int]Partition

type Partition interface {
	Start() uint64
	End() uint64
	Id() string
	Type() string
	Name() string
	Number() uint
	Size() uint64
}

type UdevInfo struct {
	Name       string
	SysPath    string
	Symlinks   []string
	Properties map[string]string
}

type FreeSpace struct {
	Start uint64
	End   uint64
}

func (f *FreeSpace) Size() uint64 {
	return f.End - f.Start
}

func ScanSystem() (DiskSet, error) {
	return nil, nil
}

func ScanDisk(devicePath string) (Disk, error) {
	return nil, nil
}

type DiskSetByType map[DiskType]DiskSet

func ScanDiskByType(diskType DiskType) (DiskSet, error) {
	return nil, nil
}
