// +build linux

package linux

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"unicode/utf16"

	"github.com/anuvu/disko"
	"github.com/rekby/gpt"
	"github.com/rekby/mbr"
	"golang.org/x/sys/unix"
)

const (
	sectorSize512 = 512
	sectorSize4k  = 4096
)

// ErrNoPartitionTable is returned if there is no partition table.
var ErrNoPartitionTable error = errors.New("no Partition Table Found")

// toGPTPartition - convert the Partition type into a gpt.Partition
func toGPTPartition(p disko.Partition, sectorSize uint) gpt.Partition {
	return gpt.Partition{
		Type:          gpt.PartType(p.Type),
		Id:            gpt.Guid(p.ID),
		FirstLBA:      Floor(p.Start, uint64(sectorSize)),
		LastLBA:       Floor(p.Last, uint64(sectorSize)),
		Flags:         gpt.Flags{},
		PartNameUTF16: getPartName(p.Name),
		TrailingBytes: []byte{},
	}
}

// getDiskType(udInfo) return the diskType for the disk represented
//   by the udev info provided.  Supports a block device
func getDiskType(udInfo disko.UdevInfo) (disko.DiskType, error) {
	var kname = udInfo.Name

	if strings.HasPrefix(kname, "nvme") {
		return disko.NVME, nil
	}

	if isKvm() {
		psuedoSsd := regexp.MustCompile("^ssd[0-9-]")
		if psuedoSsd.MatchString(udInfo.Properties["ID_SERIAL"]) {
			return disko.SSD, nil
		}
	}

	bd, err := getPartitionsBlockDevice(path.Join("/dev", kname))
	if err != nil {
		return disko.HDD, nil
	}

	syspath, err := getSysPathForBlockDevicePath(bd)
	if err != nil {
		return disko.HDD, nil
	}

	content, err := ioutil.ReadFile(
		fmt.Sprintf("%s/%s", syspath, "queue/rotational"))
	if err != nil {
		return disko.HDD,
			fmt.Errorf("failed to read %s/queue/rotational for %s", syspath, kname)
	}

	if string(content) == "0\n" {
		return disko.SSD, nil
	}

	return disko.HDD, nil
}

func getAttachType(udInfo disko.UdevInfo) disko.AttachmentType {
	bus := udInfo.Properties["ID_BUS"]
	attach := disko.UnknownAttach

	switch bus {
	case "ata":
		attach = disko.ATA
	case "usb":
		attach = disko.USB
	case "scsi":
		attach = disko.SCSI
	case "virtio":
		attach = disko.VIRTIO
	case "":
		if strings.Contains(udInfo.SysPath, "/virtio") {
			attach = disko.VIRTIO
		} else if strings.Contains(udInfo.SysPath, "/nvme/") {
			attach = disko.PCIE
		}
	}

	return attach
}

func readTableSearch(fp io.ReadSeeker, sizes []uint) (gpt.Table, uint, error) {
	const noGptFound = "Bad GPT signature"
	var gptTable gpt.Table
	var err error
	var size uint

	for _, size = range sizes {
		// consider seek failure to be fatal
		if _, err := fp.Seek(int64(size), io.SeekStart); err != nil {
			return gpt.Table{}, size, err
		}

		if gptTable, err = gpt.ReadTable(fp, uint64(size)); err != nil {
			if err.Error() == noGptFound {
				continue
			}

			return gpt.Table{}, size, err
		}

		return gptTable, size, nil
	}

	return gpt.Table{}, size, ErrNoPartitionTable
}

func readTable(fp io.ReadSeeker) (gpt.Table, uint, error) {
	return readTableSearch(fp, []uint{sectorSize512, sectorSize4k})
}

func findPartitions(fp io.ReadSeeker) (disko.PartitionSet, uint, error) {
	var err error
	var ssize uint
	var gptTable gpt.Table

	parts := disko.PartitionSet{}

	gptTable, ssize, err = readTable(fp)
	if err != nil {
		return parts, ssize, ErrNoPartitionTable
	}

	ssize64 := uint64(ssize)

	for n, p := range gptTable.Partitions {
		if p.IsEmpty() {
			continue
		}

		part := disko.Partition{
			Start:  p.FirstLBA * ssize64,
			Last:   p.LastLBA*ssize64 + ssize64 - 1,
			ID:     disko.GUID(p.Id),
			Type:   disko.PartType(p.Type),
			Name:   p.Name(),
			Number: uint(n + 1),
		}
		parts[part.Number] = part
	}

	return parts, ssize, nil
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

func getPartName(s string) [72]byte {
	codes := utf16.Encode([]rune(s))
	b := [72]byte{}

	for i, r := range codes {
		b[i*2] = byte(r)
		b[i*2+1] = byte(r >> 8) //nolint:gomnd
	}

	return b
}

// addPartitionSet - open the disk, add partitions.
//     Caller's responsibility to udevSettle
func addPartitionSet(d disko.Disk, pSet disko.PartitionSet) error {
	fp, err := os.OpenFile(d.Path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer fp.Close()

	if err := syscall.Flock(int(fp.Fd()), unix.LOCK_EX); err != nil {
		return fmt.Errorf("failed to lock %s: %s", d.Path, err)
	}

	gptTable, _, err := readTableSearch(fp, []uint{d.SectorSize})
	if err == ErrNoPartitionTable {
		gptTable, err = writeNewGPTTable(fp, d.SectorSize, d.Size)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	for _, p := range pSet {
		gptTable.Partitions[p.Number-1] = toGPTPartition(p, d.SectorSize)
	}

	_, err = writeGPTTable(fp, gptTable)

	if err != nil {
		return err
	}

	// close the file handle, releasing the lock before calling udevSettle
	// https://systemd.io/BLOCK_DEVICE_LOCKING/
	return fp.Close()
}

// writeProtectiveMBR - add a ProtectiveMBR spanning the disk.
func writeProtectiveMBR(fp io.ReadWriteSeeker, sectorSize uint, diskSize uint64) error {
	buf := make([]byte, sectorSize)

	if _, err := fp.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if _, err := io.ReadFull(fp, buf); err != nil {
		return err
	}

	m, err := newProtectiveMBR(buf, sectorSize, diskSize)
	if err != nil {
		return err
	}

	if _, err := fp.Seek(0, io.SeekStart); err != nil {
		return err
	}

	return m.Write(fp)
}

func writeNewGPTTable(fp io.ReadWriteSeeker, sectorSize uint, diskSize uint64) (gpt.Table, error) {
	guid := disko.GenGUID()
	gptTable := newGPTTable(sectorSize, diskSize, guid)

	if err := writeProtectiveMBR(fp, sectorSize, diskSize); err != nil {
		return gptTable, err
	}

	return writeGPTTable(fp, gptTable)
}

func writeGPTTable(fp io.ReadWriteSeeker, table gpt.Table) (gpt.Table, error) {
	if err := table.Write(fp); err != nil {
		fmt.Fprintf(os.Stderr, "Failed write to table: %s\n", err)
		return gpt.Table{}, err
	}

	if err := table.CreateOtherSideTable().Write(fp); err != nil {
		fmt.Fprintf(os.Stderr, "Failed write other side table: %s\n", err)
		return gpt.Table{}, err
	}

	if _, err := fp.Seek(int64(table.Header.HeaderCopyStartLBA), io.SeekStart); err != nil {
		return gpt.Table{}, err
	}

	if _, err := fp.Seek(
		int64(table.Header.HeaderStartLBA*table.SectorSize),
		io.SeekStart); err != nil {
		return gpt.Table{}, err
	}

	return gpt.ReadTable(io.ReadSeeker(fp), table.SectorSize)
}

// newProtectiveMBR - return a Protective MBR for the
// pull request to upstream mbr at https://github.com/rekby/mbr/pull/2
func newProtectiveMBR(buf []byte, sectorSize uint, diskSize uint64) (mbr.MBR, error) {
	if len(buf) < int(sectorSize) {
		return mbr.MBR{},
			fmt.Errorf("buffer too small. Must be sectorSize(%d)", sectorSize)
	}

	// error is ignored here but checked below
	myMBR, _ := mbr.Read(bytes.NewReader(buf))

	myMBR.FixSignature()

	pt := myMBR.GetPartition(1)
	pt.SetType(mbr.PART_GPT)
	pt.SetLBAStart(1)
	// Upstream pull request would set this to '- 1', not '- 2' as
	// is commonly written by linux partitioners although actually outside spec.
	pt.SetLBALen(uint32(diskSize/uint64(sectorSize)) - 2) // nolint: gomnd

	for pnum := 2; pnum <= 4; pnum++ {
		pt := myMBR.GetPartition(pnum)
		pt.SetType(mbr.PART_EMPTY)
		pt.SetLBAStart(0)
		pt.SetLBALen(0)
	}

	return *myMBR, myMBR.Check()
}

// newGPTTable - return a valid empty Table for given sectorSize and diskSize
//  pull request for gpt inclusion at https://github.com/rekby/gpt/pull/6
func newGPTTable(sectorSize32 uint, diskSize uint64, diskGUID disko.GUID) gpt.Table {
	// CreateTableForNewdiskSize will update HeaderCopyStartLBA, LastUsableLBA, and CRC
	numParts := 128
	var sectorSize = uint64(sectorSize32)
	var standardHeaderSize uint32 = 92   // Size of standard GPT-header in bytes
	var standardPartitionEntrySize = 128 // Size of standard GPT-partition entry in bytes
	var firstLBA uint64 = 34

	return gpt.Table{
		SectorSize: sectorSize,
		Header: gpt.Header{
			Signature:               [8]byte{0x45, 0x46, 0x49, 0x20, 0x50, 0x41, 0x52, 0x54},
			Revision:                0x10000, // nolint:gomnd
			Size:                    standardHeaderSize,
			CRC:                     0,
			Reserved:                0,
			HeaderStartLBA:          1,
			HeaderCopyStartLBA:      0,
			FirstUsableLBA:          firstLBA, // nolint:gomnd
			LastUsableLBA:           0,
			DiskGUID:                gpt.Guid(diskGUID),
			PartitionsTableStartLBA: 2, // nolint:gomnd
			PartitionsArrLen:        uint32(numParts),
			PartitionEntrySize:      uint32(standardPartitionEntrySize),
			PartitionsCRC:           0,
			TrailingBytes:           make([]byte, sectorSize-uint64(standardHeaderSize)),
		},
		Partitions: make([]gpt.Partition, numParts),
	}.CreateTableForNewDiskSize(diskSize / sectorSize)
}
