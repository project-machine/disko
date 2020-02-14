// +build linux

package linux

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/anuvu/disko"
	"github.com/rekby/gpt"
)

const (
	sectorSize512 = 512
	sectorSize4k  = 4096
)

// ErrNoPartitionTable is returned if there is no partition table.
var ErrNoPartitionTable error = errors.New("no Partition Table Found")

// getDiskType(udInfo) return the diskType for the disk represented
//   by the udev info provided.  Supports a block device
func getDiskType(udInfo disko.UdevInfo) (disko.DiskType, error) {
	var kname = udInfo.Name

	if strings.HasPrefix("nvme", kname) {
		return disko.NVME, nil
	}

	if isKvm() {
		psuedoSsd := regexp.MustCompile("^ssd[0-9-]")
		if psuedoSsd.MatchString(udInfo.Properties["ID_SERIAL"]) {
			return disko.SSD, nil
		}
	}

	bd, err := getPartitionsBlockDevice(kname)
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
