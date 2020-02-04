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

func findPartitions(fp io.ReadSeeker) (disko.PartitionSet, uint, error) {
	var err error

	parts := disko.PartitionSet{}
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

		part := disko.Partition{
			Start:  p.FirstLBA * ssize,
			End:    p.LastLBA*ssize + ssize - 1,
			ID:     p.Id.String(),
			Type:   p.Type.String(),
			Name:   p.Name(),
			Number: uint(n + 1),
		}
		parts[part.Number] = part

		if p.LastLBA > max {
			max = p.LastLBA
		}
	}

	return parts, uint(ssize), nil
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
