package disko

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/rekby/gpt"
)

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
	// UNKNOWN - indicates an unknown attachment.
	UnknownAttach = iota

	// RAID - indicates that the device is attached to RAID card
	RAID

	// SCSI - indicates device is attached to scsi, but not a RAID card.
	SCSI

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

const (
	sectorSize512 = 512
	sectorSize4k  = 4096
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

	// Attachment returns the type of storage card this disk is attached to.
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
type PartitionSet map[uint]Partition

// Partition interface wraps the disk partition information.
type Partition interface {
	// Start returns the start offset of the disk partition.
	Start() uint64

	// End returns the end offset of the disk partition.
	End() uint64

	// Id returns the partition id.
	ID() string

	// Type returns the partition type.
	Type() string

	// Name returns the name of this partition.
	Name() string

	// Number returns the number of this partition.
	Number() uint

	// Size returns the size of this partition.
	Size() uint64
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

// ScanSystemDisks scans the system for all the disks and returns a disk set with
// the disk information of all the discovered disks.
func ScanSystemDisks() (DiskSet, error) {
	var err error
	var dpaths = []string{}

	names, err := getDiskNames()
	if err != nil {
		return DiskSet{}, err
	}

	for _, name := range names {
		dpath := path.Join("/dev", name)

		f, err := os.Open(dpath)
		if err != nil {
			// ENOMEDIUM will occur on a empty sd reader.
			if e, ok := err.(*os.PathError); ok {
				if e.Err == syscall.ENOMEDIUM {
					continue
				}
			}

			log.Printf("Skipping device %s: %v", name, err)

			continue
		}

		f.Close()

		dpaths = append(dpaths, dpath)
	}

	return ScanDisks(dpaths)
}

// ScanDisks scans the provided list of paths
func ScanDisks(dpaths []string) (DiskSet, error) {
	var disk Disk
	var err error
	var disks = DiskSet{}

	for _, dpath := range dpaths {
		if disk, err = ScanDisk(dpath); err != nil {
			return disks, err
		}

		disks[disk.Name()] = disk
	}

	return disks, err
}

// ScanDisk will scan for the disk specified by the disk path. Returns the disk
// information if found else error.
func ScanDisk(devicePath string) (Disk, error) {
	var err error
	var blockdev = true
	var ssize uint = sectorSize512

	name, err := getKnameForBlockDevicePath(devicePath)

	if err != nil {
		name = path.Base(devicePath)
		blockdev = false
	} else {
		bss, err := getBlockDevSize(devicePath)
		if err != nil {
			return &diskImpl{}, nil
		}
		ssize = uint(bss)
	}

	udInfo, err := GetUdevInfo(name)
	if err != nil {
		return &diskImpl{}, err
	}

	diskType, err := getDiskType(udInfo)
	if err != nil {
		return &diskImpl{}, err
	}

	disk := diskImpl{
		iName:       name,
		iPath:       devicePath,
		iSectorSize: ssize,
		iUdevInfo:   udInfo,
		iType:       diskType,
		iAttachment: getAttachType(udInfo),
	}

	fh, err := os.Open(devicePath)
	if err != nil {
		return disk, err
	}
	defer fh.Close()

	size, err := getFileSize(fh)
	if err != nil {
		return disk, err
	}

	disk.iSize = size
	parts, ssize, err := findPartitions(fh)

	if err == ErrNoPartitionTable {
		return disk, nil
	}

	if ssize != disk.iSectorSize {
		if blockdev {
			return disk, fmt.Errorf(
				"Disk %s has sector size %d and partition table sector size %d",
				disk.iPath, disk.iSectorSize, ssize)
		}

		disk.iSectorSize = ssize
	}

	disk.iPartitions = parts

	return disk, nil
}

// getDiskType(udInfo) return the diskType for the disk represented
//   by the udev info provided.  Supports a block device
func getDiskType(udInfo UdevInfo) (DiskType, error) {
	var kname = udInfo.Name

	if strings.HasPrefix("nvme", kname) {
		return NVME, nil
	}

	if isKvm() {
		psuedoSsd := regexp.MustCompile("^ssd[0-9-]")
		if psuedoSsd.MatchString(udInfo.Properties["ID_SERIAL"]) {
			return SSD, nil
		}
	}

	bd, err := getPartitionsBlockDevice(kname)
	if err != nil {
		return HDD, nil
	}

	syspath, err := getSysPathForBlockDevicePath(bd)
	if err != nil {
		return HDD, nil
	}

	content, err := ioutil.ReadFile(
		fmt.Sprintf("%s/%s", syspath, "queue/rotational"))
	if err != nil {
		return HDD,
			fmt.Errorf("failed to read %s/queue/rotational for %s", syspath, kname)
	}

	if string(content) == "0\n" {
		return SSD, nil
	}

	return HDD, nil
}

func getAttachType(udInfo UdevInfo) AttachmentType {
	var bus string = udInfo.Properties["ID_BUS"]
	var attach AttachmentType = UnknownAttach

	switch bus {
	case "ata":
		attach = ATA
	case "usb":
		attach = USB
	case "scsi":
		attach = SCSI
	case "virtio":
		attach = VIRTIO
	case "":
		if strings.Contains(udInfo.Properties["DEVPATH"], "virtio") {
			attach = VIRTIO
		} else if strings.HasPrefix(udInfo.Name, "nvme") {
			attach = PCIE
		}
	}

	return attach
}

func findPartitions(fp io.ReadSeeker) (PartitionSet, uint, error) {
	var err error

	parts := PartitionSet{}
	ssize := uint64(sectorSize512)
	size4k := uint64(sectorSize4k)

	if _, err := fp.Seek(int64(ssize), io.SeekStart); err != nil {
		return parts, uint(ssize), err
	}

	var gptTable gpt.Table
	var noGptFound = "Bad GPT signature"

	gptTable, err = gpt.ReadTable(fp, ssize)
	if err != nil {
		if err.Error() != noGptFound {
			return parts, uint(ssize), err
		}

		// No GPT with 512, try for a 4096 byte sector size.
		var err4k error

		if _, err4k = fp.Seek(int64(size4k), io.SeekStart); err4k == nil {
			gptTable, err4k = gpt.ReadTable(fp, size4k)
		}

		if err4k != nil {
			if err4k.Error() == noGptFound {
				return parts, uint(ssize), ErrNoPartitionTable
			}

			return parts, uint(size4k), err4k
		}

		ssize = size4k
	}

	max := gptTable.Header.FirstUsableLBA

	for n, p := range gptTable.Partitions {
		if p.IsEmpty() {
			continue
		}

		part := partitionImpl{
			iStart:  p.FirstLBA * ssize,
			iEnd:    p.LastLBA*ssize + ssize - 1,
			iID:     p.Id.String(),
			iType:   p.Type.String(),
			iName:   p.Name(),
			iNumber: uint(n + 1),
		}
		parts[part.iNumber] = part

		if p.LastLBA > max {
			max = p.LastLBA
		}
	}

	return parts, uint(ssize), nil
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

// ErrNoPartitionTable is returned if there is no partition table.
var ErrNoPartitionTable error = errors.New("no Partition Table Found")

type diskImpl struct {
	iName       string
	iPath       string
	iSize       uint64
	iSectorSize uint
	iType       DiskType
	iAttachment AttachmentType
	iPartitions PartitionSet
	iUdevInfo   UdevInfo
}

func (d diskImpl) Name() string {
	return d.iName
}

func (d diskImpl) Path() string {
	return d.iPath
}

func (d diskImpl) Size() uint64 {
	return d.iSize
}

func (d diskImpl) SectorSize() uint {
	return d.iSectorSize
}

func (d diskImpl) Type() DiskType {
	return d.iType
}

func (d diskImpl) Attachment() AttachmentType {
	return d.iAttachment
}

func (d diskImpl) Partitions() PartitionSet {
	return d.iPartitions
}

func (d diskImpl) UdevInfo() UdevInfo {
	return d.iUdevInfo
}

func (d *diskImpl) FreeSpacesWithMin(minSize uint64) []FreeSpace {
	// Stay out of the first 1Mebibyte
	// Leave 33 sectors at end (for GPT second header) and round 1MiB down.
	end := ((d.Size() - uint64(d.SectorSize())*33) / Mebibyte) * Mebibyte
	used := []uRange{{0, 1*Mebibyte - 1}, {end, d.Size()}}
	avail := []FreeSpace{}

	for _, p := range d.Partitions() {
		used = append(used, uRange{p.Start(), p.End()})
	}

	for _, g := range findRangeGaps(used, 0, d.Size()) {
		if g.Size() < minSize {
			continue
		}

		avail = append(avail, FreeSpace(g))
	}

	return avail
}

func (d diskImpl) FreeSpace() []FreeSpace {
	return d.FreeSpacesWithMin(ExtentSize)
}

func (d diskImpl) CreatePartition(part Partition) error {
	return nil
}

func (d diskImpl) DeletePartition(partNum int) error {
	return nil
}

func (d diskImpl) Wipe() error {
	return nil
}

func (d diskImpl) String() string {
	var avail uint64 = 0

	fs := d.FreeSpace()

	for _, f := range d.FreeSpace() {
		avail += f.Size()
	}

	mbsize := func(n uint64) string {
		if (n)%Mebibyte == 0 {
			return fmt.Sprintf("%dMiB", (n)/Mebibyte)
		}

		return fmt.Sprintf("%d", n)
	}

	return fmt.Sprintf(
		"%s (%s) Size=%s NumParts=%d FreeSpace=%s/%d, SectorSize=%d Attachment=%s Type=%s",
		d.iName, d.iPath, mbsize(d.iSize), len(d.iPartitions),
		mbsize(avail), len(fs), d.iSectorSize,
		string(d.iAttachment), string(d.iType))
}

func (d diskImpl) Details() string {
	fss := d.FreeSpace()
	var fsn int = 0

	mbsize := func(n, o uint64) string {
		if (n+o)%Mebibyte == 0 {
			return fmt.Sprintf("%d MiB", (n+o)/Mebibyte)
		}

		return fmt.Sprintf("%d", n)
	}

	mbo := func(n uint64) string { return mbsize(n, 0) }
	mbe := func(n uint64) string { return mbsize(n, 1) }
	lfmt := "[%2s  %10s %10s %10s %-16s]\n"
	buf := fmt.Sprintf(lfmt, "#", "Start", "End", "Size", "Name")

	for _, p := range d.Partitions() {
		if fsn < len(fss) && fss[fsn].Start < p.Start() {
			buf += fmt.Sprintf(lfmt, "-", mbo(fss[fsn].Start), mbe(fss[fsn].End), mbo(fss[fsn].Size()), "<free>")
			fsn++
		}

		buf += fmt.Sprintf(lfmt,
			fmt.Sprintf("%d", p.Number()), mbo(p.Start()), mbe(p.End()), mbo(p.Size()), p.Name())
	}

	if fsn < len(fss) {
		buf += fmt.Sprintf(lfmt, "-", mbo(fss[fsn].Start), mbe(fss[fsn].End), mbo(fss[fsn].Size()), "<free>")
	}

	return buf
}

type partitionImpl struct {
	iStart  uint64
	iEnd    uint64
	iID     string
	iType   string
	iName   string
	iNumber uint
}

func (p partitionImpl) Start() uint64 {
	return p.iStart
}

func (p partitionImpl) End() uint64 {
	return p.iEnd
}

func (p partitionImpl) ID() string {
	return p.iID
}

func (p partitionImpl) Type() string {
	return p.iType
}

func (p partitionImpl) Name() string {
	return p.iName
}

func (p partitionImpl) Number() uint {
	return p.iNumber
}

func (p partitionImpl) Size() uint64 {
	return p.End() - p.Start() + 1
}

func getDiskNames() ([]string, error) {
	realDiskKnameRegex := regexp.MustCompile("^((s|v|xv|h)d[a-z]|nvme[0-9]n[0-9]+)$")
	disks := []string{}

	files, err := ioutil.ReadDir("/sys/block")
	if err != nil {
		return []string{}, err
	}

	for _, file := range files {
		if realDiskKnameRegex.MatchString(file.Name()) {
			disks = append(disks, file.Name())
		}
	}

	return disks, nil
}

func getKnameForBlockDevicePath(dev string) (string, error) {
	// given '/dev/sda1' (or any valid block device path) return 'sda'
	kname, err := getSysPathForBlockDevicePath(dev)
	if err != nil {
		return "", err
	}

	return path.Base(kname), nil
}

func getSysPathForBlockDevicePath(dev string) (string, error) {
	// Return the path in /sys/class/block/<device> for a given
	// block device kname or path.
	var syspath string
	var sysdir string = "/sys/class/block"

	if strings.Contains(dev, "/") {
		// after symlink resolution, devpath = '/dev/sda' or '/dev/sdb1'
		// no longer something like /dev/disk/by-id/foo
		devpath, err := filepath.EvalSymlinks(dev)
		if err != nil {
			return "", err
		}

		syspath = fmt.Sprintf("%s/%s", sysdir, path.Base(devpath))
	} else {
		// assume this is 'sda', something that would be in /sys/class/block
		syspath = fmt.Sprintf("%s/%s", sysdir, dev)
	}

	_, err := os.Stat(syspath)
	if err != nil {
		return "", err
	}

	return syspath, nil
}

func getPartitionsBlockDevice(dev string) (string, error) {
	// return the block device name ('sda') given input
	// of 'sda1', /dev/sda1, or /dev/sda
	syspath, err := getSysPathForBlockDevicePath(dev)
	if err != nil {
		return "", err
	}

	_, err = ioutil.ReadFile(fmt.Sprintf("%s/%s", syspath, "partition"))
	if err != nil {
		// this is a block device itself, no /sys/class/block/<dev>/partition
		return filepath.EvalSymlinks(dev)
	}

	// evalSymlinks on a partition will return
	// /sys/devices/<bus>/<path>/<compoents>/block/<diskName>/<PartitionName>
	sysfull, err := filepath.EvalSymlinks(syspath)
	if err != nil {
		return "", err
	}

	return path.Base(path.Dir(sysfull)), nil
}
