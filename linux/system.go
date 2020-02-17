// +build linux

package linux

import (
	"fmt"
	"log"
	"os"
	"path"
	"syscall"

	"github.com/anuvu/disko"
)

type linuxSystem struct {
}

// System returns an linux specific implementation of disko.System interface.
func System() disko.System {
	return &linuxSystem{}
}

func (ls *linuxSystem) ScanAllDisks(filter disko.DiskFilter) (disko.DiskSet, error) {
	var err error
	var dpaths = []string{}

	names, err := getDiskNames()
	if err != nil {
		return disko.DiskSet{}, err
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

	return ls.ScanDisks(filter, dpaths...)
}

func (ls *linuxSystem) ScanDisks(filter disko.DiskFilter,
	dpaths ...string) (disko.DiskSet, error) {
	disks := disko.DiskSet{}

	for _, dpath := range dpaths {
		disk, err := ls.ScanDisk(dpath)
		if err != nil {
			return disks, err
		}

		if filter(disk) {
			// Accepted so add to the set
			disks[disk.Name] = disk
		}
	}

	return disks, nil
}

func (ls *linuxSystem) ScanDisk(devicePath string) (disko.Disk, error) {
	var err error
	var blockdev = true
	var ssize uint = sectorSize512

	name, err := getKnameForBlockDevicePath(devicePath)

	if err != nil {
		name = path.Base(devicePath)
		blockdev = false
	} else {
		bss, err := getBlockSize(devicePath)
		if err != nil {
			return disko.Disk{}, nil
		}
		ssize = uint(bss)
	}

	udInfo, err := GetUdevInfo(name)
	if err != nil {
		return disko.Disk{}, err
	}

	diskType, err := getDiskType(udInfo)
	if err != nil {
		return disko.Disk{}, err
	}

	disk := disko.Disk{
		Name:       name,
		Path:       devicePath,
		SectorSize: ssize,
		UdevInfo:   udInfo,
		Type:       diskType,
		Attachment: getAttachType(udInfo),
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

	disk.Size = size
	parts, ssize, err := findPartitions(fh)

	if err == ErrNoPartitionTable {
		return disk, nil
	}

	if ssize != disk.SectorSize {
		if blockdev {
			return disk, fmt.Errorf(
				"disk %s has sector size %d and partition table sector size %d",
				disk.Path, disk.SectorSize, ssize)
		}

		disk.SectorSize = ssize
	}

	disk.Partitions = parts

	return disk, nil
}

func (ls *linuxSystem) CreatePartition(d disko.Disk, p disko.Partition) error {
	if err := addPartitionSet(d, disko.PartitionSet{p.Number: p}); err != nil {
		return err
	}

	return udevSettle()
}

func (ls *linuxSystem) DeletePartition(d disko.Disk, number uint) error {
	if err := deletePartitions(d, []uint{number}); err != nil {
		return err
	}

	return udevSettle()
}

func (ls *linuxSystem) Wipe(d disko.Disk) error {
	return nil
}
