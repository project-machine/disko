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
	"github.com/anuvu/disko/partid"
	"github.com/rekby/gpt"
	"github.com/rekby/mbr"
	"golang.org/x/sys/unix"
)

const (
	sectorSize512 = 512
	sectorSize4k  = 4096
	max32         = 0xFFFFFFFF
)

// ErrNoPartitionTable is returned if there is no partition table.
var ErrNoPartitionTable = errors.New("no Partition Table Found")

var xenbusSysPathMatch = regexp.MustCompile(`/devices/vbd-\d+/block/`)

// nolint: gochecknoglobals
var emptyGUID = disko.GUID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

// toGPTPartition - convert the Partition type into a gpt.Partition
func toGPTPartition(p disko.Partition, sectorSize uint) gpt.Partition {
	if p.ID == emptyGUID {
		p.ID = disko.GenGUID()
	}

	return gpt.Partition{
		Type:          gpt.PartType(p.Type),
		Id:            gpt.Guid(p.ID),
		FirstLBA:      Floor(p.Start, uint64(sectorSize)) / uint64(sectorSize),
		LastLBA:       Floor(p.Last, uint64(sectorSize)) / uint64(sectorSize),
		Flags:         gpt.Flags{},
		PartNameUTF16: getPartName(p.Name),
		TrailingBytes: []byte{},
	}
}

// gptToDiskoPartition - convert the gpt.Partition type to a disko.Partition
func gptToDiskoPartition(p gpt.Partition, num uint, sectorSize uint) disko.Partition {
	// this is lossy on TrailingBytes  and Flags :-(
	return disko.Partition{
		Start:  p.FirstLBA * uint64(sectorSize),
		Last:   p.LastLBA*uint64(sectorSize) + uint64(sectorSize-1),
		ID:     disko.GUID(p.Id),
		Type:   disko.PartType(p.Type),
		Name:   p.Name(),
		Number: num,
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
		if strings.HasPrefix(udInfo.Properties["ID_SERIAL"], "ssd-") {
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
		} else if strings.Contains(udInfo.SysPath, "/virtual/block/nbd") {
			attach = disko.NBD
		} else if strings.Contains(udInfo.SysPath, "/virtual/block/loop") {
			attach = disko.LOOP
		} else if xenbusSysPathMatch.MatchString(udInfo.SysPath) {
			attach = disko.XENBUS
		}
	}

	return attach
}

func readGPTTableSearch(fp io.ReadSeeker, sizes []uint) (gpt.Table, uint, error) {
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

func readGPTTable(fp io.ReadSeeker) (gpt.Table, uint, error) {
	return readGPTTableSearch(fp, []uint{sectorSize512, sectorSize4k})
}

func readMBRTable(fp io.ReadSeeker) (disko.PartitionSet, error) {
	parts := disko.PartitionSet{}

	if _, err := fp.Seek(0, io.SeekStart); err != nil {
		return parts, err
	}

	mbrTable, err := mbr.Read(fp)

	if err == mbr.ErrorBadMbrSign {
		return parts, ErrNoPartitionTable
	}

	for i, p := range mbrTable.GetAllPartitions() {
		if p.IsEmpty() {
			continue
		}

		buf := [16]byte{}
		buf[15] = byte(p.GetType())

		part := disko.Partition{
			Start:  uint64(p.GetLBAStart()) * sectorSize512,
			Last:   uint64(p.GetLBALast())*sectorSize512 + sectorSize512 - 1,
			Type:   disko.PartType(buf),
			Number: uint(i + 1),
		}
		parts[part.Number] = part
	}

	return parts, nil
}

func findPartitions(fp io.ReadSeeker) (disko.PartitionSet, disko.TableType, uint, error) {
	var err error
	var ssize uint
	var gptTable gpt.Table

	gptTable, ssize, err = readGPTTable(fp)
	if err == ErrNoPartitionTable {
		parts, err := readMBRTable(fp)
		if err == ErrNoPartitionTable {
			return parts, disko.TableNone, ssize, nil
		}

		return parts, disko.MBR, sectorSize512, err
	}

	if err != nil {
		return disko.PartitionSet{}, disko.GPT, ssize, err
	}

	parts := disko.PartitionSet{}
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

	return parts, disko.GPT, ssize, nil
}

func getDiskNames() ([]string, error) {
	realDiskKnameRegex := regexp.MustCompile("^((s|v|xv|h)d[a-z]|nvme[0-9]n[0-9]|mmcblk[0-9]+)$")
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

func getPathForKname(kname string) string {
	return path.Join("/dev", kname)
}

func getKnameAndPathForBlockDevice(nameOrPath string) (string, string, error) {
	syspath, err := getSysPathForBlockDevicePath(nameOrPath)
	if err != nil {
		return "", "", err
	}

	kname := path.Base(syspath)

	return kname, getPathForKname(kname), nil
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
	var sysdir = "/sys/class/block"

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
		// dev is a block device, there is no /sys/class/block/<dev>/partition
		return path.Base(syspath), nil
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
		b[i*2+1] = byte(r >> 8) // nolint:gomnd
	}

	return b
}

func wipeDisk(disk disko.Disk) error {
	fp, err := os.OpenFile(disk.Path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer fp.Close()

	if err := zeroStartEnd(fp, int64(0), int64(disk.Size)); err != nil {
		return err
	}

	for _, p := range disk.Partitions {
		// The point of this operation is to wipe.  Avoid out of range errors
		// that could happen as part of a bad partition table.
		end := p.Last
		if end > disk.Size {
			end = disk.Size
		}

		if err := zeroStartEnd(fp, int64(p.Start), int64(end)); err != nil {
			return err
		}
	}

	return nil
}

// zeroStartEnd - zero the start and end provided with 1MiB bytes of zeros.
func zeroStartEnd(fp io.WriteSeeker, start int64, last int64) error {
	if last <= start {
		return fmt.Errorf("last %d < start %d", last, start)
	}

	wlen := int64(disko.Mebibyte)
	bufZero := make([]byte, wlen)

	// 3 cases.
	// a.) start + wlen < last - wlen (two full writes)
	// b.) start + wlen >= last (one possibly short write)
	// c.) start + wlen >= last - wlen (overlapping zero ranges)
	type ws struct{ start, size int64 }
	var writes = []ws{{start, wlen}, {last - wlen, wlen}}
	var wnum int
	var err error

	if start+wlen >= last {
		writes = []ws{{start, last - start}}
	} else if start+wlen >= last-wlen {
		writes = []ws{{start, wlen}, {start + wlen, last - (start + wlen)}}
	}

	for _, w := range writes {
		if _, err = fp.Seek(w.start, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek to %d to write %v", w.start, w)
		}

		wnum, err = fp.Write(bufZero[:w.size])
		if err != nil {
			return fmt.Errorf("failed to write %v", w)
		}

		if int64(wnum) != w.size {
			return fmt.Errorf("wrote only %d bytes of %v", wnum, w)
		}
	}

	return nil
}

func addPartitionSetMBR(fp io.ReadWriteSeeker, d disko.Disk, pSet disko.PartitionSet) error {
	if err := rangeCheckParts(d, pSet); err != nil {
		return err
	}

	if _, err := fp.Seek(0, io.SeekStart); err != nil {
		return err
	}

	mbrTable, err := mbr.Read(fp)
	if err == mbr.ErrorBadMbrSign {
		// the Read(0) does call Check(), but only returns the first error. That may be fixed
		// by FixingSignature, but need to check if that fixes everything.
		mbrTable.FixSignature()

		if err := mbrTable.Check(); err != nil {
			return err
		}
	} else {
		return err
	}

	for _, p := range pSet {
		mPart := mbrTable.GetPartition(int(p.Number))
		mPart.SetLBAStart(uint32(p.Start) / uint32(d.SectorSize))
		mPart.SetLBALen(uint32(p.Size()) / uint32(d.SectorSize))
		mType, err := partid.PartTypeToMBR(p.Type)

		if err != nil {
			return err
		}

		mPart.SetType(mbr.PartitionType(mType))

		if err := zeroStartEnd(fp, int64(p.Start), int64(p.Last)); err != nil {
			return fmt.Errorf("failed to zero partition %d: %s", p.Number, err)
		}
	}

	if _, err := fp.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if err := mbrTable.Write(fp); err != nil {
		return err
	}

	return nil
}

func updatePartitionSetGPT(fp io.ReadWriteSeeker, d disko.Disk, pSet disko.PartitionSet) error {
	gptTable, _, err := readGPTTableSearch(fp, []uint{d.SectorSize})
	if err != nil {
		return err
	}

	newParts := disko.PartitionSet{}

	for n, p := range pSet {
		gPart := gptTable.Partitions[n-1]
		if gPart.IsEmpty() {
			return fmt.Errorf("cannot update disk %s.Path partition %d: partition does not exist",
				d.Path, p.Number)
		}

		newPt := gptToDiskoPartition(gPart, n, d.SectorSize)

		// Only the GUID, Type and Name get updated.
		if newPt.ID != emptyGUID {
			newPt.ID = p.ID
		}

		if newPt.Type != partid.Empty {
			newPt.Type = p.Type
		}

		if p.Name != "" {
			newPt.Name = p.Name
		}

		newParts[n] = newPt
		gptTable.Partitions[n-1] = toGPTPartition(newPt, d.SectorSize)
	}

	if _, err := writeGPTTable(fp, gptTable, d.Size); err != nil {
		return err
	}

	return nil
}

func updatePartitionSetMBR(_ *os.File, d disko.Disk, pSet disko.PartitionSet) error {
	return fmt.Errorf("MBR partition update is not implemented. Cannot update %d partitions on %s",
		len(pSet), d.Path)
}

// nolint:scopelint // https://github.com/kyoh86/scopelint/issues/12
func updatePartitions(d disko.Disk, pSet disko.PartitionSet) error {
	err := withLockedFile(d.Path,
		func(fp *os.File, fInfo os.FileInfo) error {
			if d.Table == disko.MBR {
				return updatePartitionSetMBR(fp, d, pSet)
			} else if d.Table == disko.GPT {
				return updatePartitionSetGPT(fp, d, pSet)
			} else if d.Table == disko.TableNone {
				return fmt.Errorf("cannot update partitions on disk %s: it has no partition table",
					d.Name)
			}
			return fmt.Errorf(
				"cannot update partitions on disk %s: partition table '%s' is not supported",
				d.Name, d.Table)
		})

	if err != nil {
		return err
	}

	if err := udevSettle(); err != nil {
		return err
	}

	return genPartChangeUEvent(d, pSet)
}

func rangeCheckParts(d disko.Disk, pSet disko.PartitionSet) error {
	maxSize := d.Size

	if d.Table == disko.MBR {
		maxSize = uint64(max32) * uint64(d.SectorSize)
	}

	maxEnd := ((maxSize - uint64(d.SectorSize)*33) / disko.Mebibyte) * disko.Mebibyte
	minStart := disko.Mebibyte

	const minPartNum, maxPartNumMBR, maxPartNumGPT = 1, 4, 128

	maxPartNum := uint(maxPartNumGPT)
	if d.Table == disko.MBR {
		maxPartNum = uint(maxPartNumMBR)
	}

	for _, p := range pSet {
		if p.Number < uint(minPartNum) || p.Number > maxPartNum {
			return fmt.Errorf("partition number %d is out of range (%d-%d) for %s",
				p.Number, minPartNum, maxPartNum, d.Table)
		}

		if p.Start < minStart {
			return fmt.Errorf("partition %d start (%d) is too low. Must be >= %d",
				p.Number, p.Start, minStart)
		}

		if p.Last >= maxEnd {
			return fmt.Errorf("partition %d Last (%d) is too high. Must be < %d",
				p.Number, p.Last, maxEnd)
		}
	}

	return nil
}

func addPartitionSetGPT(fp io.ReadWriteSeeker, d disko.Disk, pSet disko.PartitionSet) error {
	gptTable, _, err := readGPTTableSearch(fp, []uint{d.SectorSize})
	if err == ErrNoPartitionTable {
		gptTable, err = writeNewGPTTable(fp, d.SectorSize, d.Size)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	maxEnd := ((d.Size - uint64(d.SectorSize)*33) / disko.Mebibyte) * disko.Mebibyte
	minStart := disko.Mebibyte

	for _, p := range pSet {
		gptTable.Partitions[p.Number-1] = toGPTPartition(p, d.SectorSize)

		if p.Start < minStart {
			return fmt.Errorf("partition %d start (%d) is too low. Must be >= %d",
				p.Number, p.Start, minStart)
		}

		if p.Last >= maxEnd {
			return fmt.Errorf("partition %d Last (%d) is too high. Must be < %d",
				p.Number, p.Last, maxEnd)
		}

		if err := zeroStartEnd(fp, int64(p.Start), int64(p.Last)); err != nil {
			return fmt.Errorf("failed to zero partition %d: %s", p.Number, err)
		}
	}

	if _, err := writeGPTTable(fp, gptTable, d.Size); err != nil {
		return err
	}

	return nil
}

// addPartitionSet - open the disk, add partitions.
//     Caller's responsibility to udevSettle
// nolint:scopelint // https://github.com/kyoh86/scopelint/issues/12
func addPartitionSet(d disko.Disk, pSet disko.PartitionSet) error {
	if d.Table != disko.MBR && d.Table != disko.GPT && d.Table != disko.TableNone {
		return fmt.Errorf("cannot add partition disk %s with table type %s", d.Name, d.Table)
	}

	if err := rangeCheckParts(d, pSet); err != nil {
		return err
	}

	// Add the devices and call kernelAddParts with a lock.  After doing so, the kernel
	// should know about the devices, but udev will not have processed any events
	// because of the lock.  After lock is given up, generate Change events.
	err := withLockedFile(d.Path, func(fp *os.File, fInfo os.FileInfo) error {
		if d.Table == disko.MBR {
			if err := addPartitionSetMBR(fp, d, pSet); err != nil {
				return err
			}
		} else {
			if err := addPartitionSetGPT(fp, d, pSet); err != nil {
				return nil
			}
		}

		if fInfo.Mode()&os.ModeDevice == 0 {
			return nil
		}

		return kernelAddParts(d, pSet)
	})

	if err != nil {
		return err
	}

	if err := udevSettle(); err != nil {
		return err
	}

	return genPartChangeUEvent(d, pSet)
}

func deletePartitionSetMBR(fp io.ReadWriteSeeker, pNums []uint) error {
	if _, err := fp.Seek(0, io.SeekStart); err != nil {
		return err
	}

	mbrTable, err := mbr.Read(fp)
	if err != nil {
		return err
	}

	for _, pNum := range pNums {
		if pNum < 1 || pNum > 4 {
			return fmt.Errorf("cannot delete partition %d from MBR. Invalid number", pNum)
		}

		pt := mbrTable.GetPartition(int(pNum))

		// pt.SetBootable(false) // https://github.com/rekby/mbr/pull/3/commits
		pt.SetType(mbr.PART_EMPTY)
		pt.SetLBAStart(0)
		pt.SetLBALen(0)
	}

	if err := mbrTable.Write(fp); err != nil {
		return err
	}

	return nil
}

func deletePartitionSetGPT(fp io.ReadWriteSeeker, d disko.Disk, pNums []uint) error {
	gptTable, _, err := readGPTTableSearch(fp, []uint{d.SectorSize})
	if err != nil {
		return err
	}

	emptyPart := toGPTPartition(
		disko.Partition{
			Start: 0,
			Last:  0,
			ID:    emptyGUID,
			Type:  partid.Empty,
		}, d.SectorSize)

	for _, pNum := range pNums {
		gptTable.Partitions[pNum-1] = emptyPart
	}

	if _, err := writeGPTTable(fp, gptTable, d.Size); err != nil {
		return err
	}

	return nil
}

// nolint:scopelint // https://github.com/kyoh86/scopelint/issues/12
func withLockedFile(path string, cb func(*os.File, os.FileInfo) error) error {
	fp, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer fp.Close()

	if err := syscall.Flock(int(fp.Fd()), unix.LOCK_EX); err != nil {
		return fmt.Errorf("failed to lock %s: %s", path, err)
	}

	info, err := fp.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat %s: %s", path, err)
	}

	if err := cb(fp, info); err != nil {
		return err
	}

	if err := fp.Sync(); err != nil {
		return err
	}

	return nil
}

// kernelDelParts - ask the kernel to remove a partition, and then remove it from /dev.
//  this can be executed with a lock.  successful 'delpart /dev/disk 3' will remove all
//  symlinks in /dev/ to /dev/disk3 even if disk is locked.
func kernelDelParts(d disko.Disk, pNums []uint) error {
	for _, pNum := range pNums {
		bPath := fmt.Sprintf("/sys/class/block/%s", GetPartitionKname(d.Name, pNum))

		// If the kernel does not know about this device, then 'delpart' will fail.
		// ignore other errors for now.
		if _, err := os.Stat(bPath); err != nil && os.IsNotExist(err) {
			continue
		}

		if err := runCommand("delpart", d.Path, fmt.Sprintf("%d", pNum)); err != nil {
			return err
		}
	}

	return nil
}

func kernelAddParts(d disko.Disk, pSet disko.PartitionSet) error {
	for _, p := range pSet {
		if err := runCommand("addpart", d.Path,
			fmt.Sprintf("%d", p.Number),
			fmt.Sprintf("%d", p.Start/sectorSize512),
			fmt.Sprintf("%d", p.Size()/sectorSize512)); err != nil {
			return err
		}
	}

	return nil
}

func genPartChangeUEvent(d disko.Disk, pSet disko.PartitionSet) error {
	if isBlk, err := blockDeviceExists(d.Path); err != nil {
		return err
	} else if !isBlk {
		return nil
	}

	for _, p := range pSet {
		uePath := fmt.Sprintf("/sys/class/block/%s/uevent", GetPartitionKname(d.Name, p.Number))
		if err := ioutil.WriteFile(uePath, []byte("change"), 0600); err != nil && os.IsNotExist(err) {
			return fmt.Errorf("%s did not exist: %v", uePath, err)
		} else if err != nil {
			return fmt.Errorf("failed to write 'change' to %s: %v", uePath, err)
		}
	}

	return nil
}

// nolint:scopelint // https://github.com/kyoh86/scopelint/issues/12
func deletePartitions(d disko.Disk, pNums []uint) error {
	return withLockedFile(d.Path, func(fp *os.File, fInfo os.FileInfo) error {
		if d.Table == disko.MBR {
			if err := deletePartitionSetMBR(fp, pNums); err != nil {
				return err
			}
		} else {
			if err := deletePartitionSetGPT(fp, d, pNums); err != nil {
				return err
			}
		}
		if fInfo.Mode()&os.ModeDevice == 0 {
			return nil
		}

		return kernelDelParts(d, pNums)
	})
}

func blockDeviceExists(bpath string) (bool, error) {
	info, err := os.Stat(bpath)

	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return info.Mode()&os.ModeDevice != 0, nil
}

// GetPartitionKname - get the name of the num partition device on diskName
func GetPartitionKname(diskName string, num uint) string {
	endsWithNum := regexp.MustCompile("[0-9]$")
	sep := ""

	if endsWithNum.MatchString(diskName) {
		sep = "p"
	}

	return fmt.Sprintf("%s%s%d", diskName, sep, num)
}

// writeProtectiveMBR - add a ProtectiveMBR spanning the disk.
// This preserves anything in the first sector that is outside of the partition table.
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
	ntArgs := gpt.NewTableArgs{
		SectorSize: uint64(sectorSize),
		DiskGuid:   gpt.Guid(disko.GenGUID())}
	gptTable := gpt.NewTable(diskSize, &ntArgs)

	return writeGPTTable(fp, gptTable, diskSize)
}

func writeGPTTable(fp io.ReadWriteSeeker, table gpt.Table, diskSize uint64) (gpt.Table, error) {
	if err := writeProtectiveMBR(fp, uint(table.SectorSize), diskSize); err != nil {
		return gpt.Table{}, err
	}

	table = table.CreateTableForNewDiskSize(diskSize / table.SectorSize)
	if _, err := fp.Seek(int64(table.SectorSize), io.SeekStart); err != nil {
		return gpt.Table{}, err
	}

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

	// https://en.wikipedia.org/wiki/Master_boot_record
	// partition table takes up 440 (0x1BE) to 511 (0x1FF).  We zero locations
	// of the partitions, and leave the rest.
	for offset, i := 0x1BE, 0; i < 16*4; i++ {
		buf[offset+i] = 0
	}
	// then explicitly write the mbr signature
	buf[0x1FE] = 0x55
	buf[0x1FF] = 0xAA

	myMBR, err := mbr.Read(bytes.NewReader(buf))
	if err != nil {
		return mbr.MBR{}, err
	}

	pt := myMBR.GetPartition(1)
	pt.SetType(mbr.PART_GPT)
	pt.SetLBAStart(1)

	// If mbr.Check did not complain, we would just always write the
	// length as 0xFFFFFFFF which is what windows and sfdisk do.
	// sfdisk actually complains about our -1 value.
	// see https://github.com/rekby/mbr/pull/2/files
	max := uint64(max32)
	if diskSize/uint64(sectorSize) > max {
		pt.SetLBALen(uint32(max) - 1)
	} else {
		pt.SetLBALen(uint32(diskSize/uint64(sectorSize)) - 1)
	}

	for pnum := 2; pnum <= 4; pnum++ {
		pt := myMBR.GetPartition(pnum)
		pt.SetType(mbr.PART_EMPTY)
		pt.SetLBAStart(0)
		pt.SetLBALen(0)
	}

	return *myMBR, myMBR.Check()
}
