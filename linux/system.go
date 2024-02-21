package linux

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
	"machinerun.io/disko"
	"machinerun.io/disko/megaraid"
	"machinerun.io/disko/smartpqi"
)

type linuxSystem struct {
	raidctrls []RAIDController
}

// System returns an linux specific implementation of disko.System interface.
func System() disko.System {
	return &linuxSystem{
		raidctrls: []RAIDController{
			megaraid.CachingStorCli(),
			smartpqi.ArcConf(),
		},
	}
}

// example below, of an azure vmbus disk that is ephemeral.
// matching intent of /lib/udev/rules.d/66-azure-ephemeral.rules
// /devices/LNXSYSTM:00/LNXSYBUS:00/PNP0A03:00/device:07/VMBUS:01/00000000-0001-8899-0000-000000000000/
//
//	host1/target1:0:1/1:0:1:0/block/sdb
var vmbusSyspathEphemeral = regexp.MustCompile(`.*/VMBUS:\d\d/00000000-0001-\d{4}-\d{4}-\d{12}/host.*`)

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

func getDiskReadOnly(kname string) (bool, error) {
	syspath, err := getSysPathForBlockDevicePath(kname)
	if err != nil {
		return false, err
	}

	syspathReadOnly := syspath + "/ro"
	content, err := ioutil.ReadFile(syspathReadOnly)

	if err != nil {
		return false, err
	}

	val := strings.TrimRight(string(content), "\n")

	if val == "1" {
		return true, nil
	} else if val == "0" {
		return false, nil
	}

	return false, fmt.Errorf("unexpected value '%s' found in %s", syspathReadOnly, val)
}

func getDiskProperties(d disko.UdevInfo) disko.PropertySet {
	props := disko.PropertySet{}

	if vmbusSyspathEphemeral.MatchString(d.SysPath) {
		props[disko.Ephemeral] = true
	}

	if d.Properties["ID_MODEL"] == "Amazon EC2 NVMe Instance Storage" {
		props[disko.Ephemeral] = true
	}

	return props
}

//nolint:funlen
func (ls *linuxSystem) ScanDisk(devicePath string) (disko.Disk, error) {
	var err error
	var blockdev = true
	var ssize uint = sectorSize512
	var diskType disko.DiskType
	var attachType disko.AttachmentType
	var ro bool

	name, err := getKnameForBlockDevicePath(devicePath)

	if err != nil {
		name = path.Base(devicePath)
		blockdev = false
	} else {
		bss, err := getBlockSize(name)
		if err != nil {
			return disko.Disk{}, err
		}
		ssize = uint(bss)
	}

	udInfo := disko.UdevInfo{}

	if blockdev {
		udInfo, err = GetUdevInfo(name)
		if err != nil {
			return disko.Disk{}, err
		}

		attachType = getAttachType(udInfo)

		for _, ctrl := range ls.raidctrls {
			if IsSysPathRAID(udInfo.Properties["DEVPATH"], ctrl.DriverSysfsPath()) {
				// we know this is device is part of a raid, so if we cannot get
				// disk type we must return an error
				dType, err := ctrl.GetDiskType(devicePath)
				if err != nil {
					return disko.Disk{}, fmt.Errorf("failed to get diskType of %q from RAID controller: %s", devicePath, err)
				}

				attachType = disko.RAID
				diskType = dType

				break
			}
		}

		// check disk type if it wasn't on raid
		if attachType != disko.RAID {
			diskType, err = getDiskType(udInfo)
			if err != nil {
				return disko.Disk{}, fmt.Errorf("error while getting disk type: %s", err)
			}
		}

		ro, err = getDiskReadOnly(name)
		if err != nil {
			return disko.Disk{}, err
		}
	} else {
		diskType = disko.TYPEFILE
		attachType = disko.FILESYSTEM

		ro = false
		if err := unix.Access(devicePath, unix.W_OK); err == unix.EACCES {
			ro = true
		} else if err != nil {
			return disko.Disk{}, err
		}
	}

	properties := getDiskProperties(udInfo)

	disk := disko.Disk{
		Name:       name,
		Path:       devicePath,
		SectorSize: ssize,
		ReadOnly:   ro,
		UdevInfo:   udInfo,
		Type:       diskType,
		Attachment: attachType,
		Properties: properties,
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
	parts, tType, ssize, err := findPartitions(fh)

	if err != nil {
		return disk, err
	}

	disk.Table = tType

	if tType == disko.GPT && ssize != disk.SectorSize {
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

func (ls *linuxSystem) CreatePartitions(d disko.Disk, pSet disko.PartitionSet) error {
	if err := addPartitionSet(d, pSet); err != nil {
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

func (ls *linuxSystem) UpdatePartition(d disko.Disk, p disko.Partition) error {
	if err := updatePartitions(d, disko.PartitionSet{p.Number: p}); err != nil {
		return err
	}

	return udevSettle()
}

func (ls *linuxSystem) UpdatePartitions(d disko.Disk, pSet disko.PartitionSet) error {
	if err := updatePartitions(d, pSet); err != nil {
		return err
	}

	return udevSettle()
}

func (ls *linuxSystem) Wipe(d disko.Disk) error {
	if err := wipeDisk(d); err != nil {
		return err
	}

	return udevSettle()
}

func (ls *linuxSystem) GetDiskType(path string, udInfo disko.UdevInfo) (disko.DiskType, error) {
	for _, ctrl := range ls.raidctrls {
		if IsSysPathRAID(udInfo.Properties["DEVPATH"], ctrl.DriverSysfsPath()) {
			dType, err := ctrl.GetDiskType(path)
			if err != nil {
				return disko.HDD, fmt.Errorf("failed to get diskType of %q from RAID controller: %s", path, err)
			}

			return dType, nil
		}
	}
	return getDiskType(udInfo)
}
