package linux

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"syscall"

	"github.com/anuvu/disko"
	"github.com/anuvu/disko/megaraid"
)

type linuxSystem struct {
	megaraid megaraid.MegaRaid
}

// System returns an linux specific implementation of disko.System interface.
func System() disko.System {
	return &linuxSystem{
		megaraid: megaraid.CachingStorCli(),
	}
}

// example below, of an azure vmbus disk that is ephemeral.
// matching intent of /lib/udev/rules.d/66-azure-ephemeral.rules
// /devices/LNXSYSTM:00/LNXSYBUS:00/PNP0A03:00/device:07/VMBUS:01/00000000-0001-8899-0000-000000000000/
//      host1/target1:0:1/1:0:1:0/block/sdb
// nolint: gochecknoglobals
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

	diskType, err := ls.getDiskType(devicePath, udInfo)
	if err != nil {
		return disko.Disk{}, err
	}

	attachType := getAttachType(udInfo)
	if megaraid.IsMegaRaidSysPath(udInfo.Properties["DEVPATH"]) {
		attachType = disko.RAID
	}

	properties := getDiskProperties(udInfo)

	disk := disko.Disk{
		Name:       name,
		Path:       devicePath,
		SectorSize: ssize,
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

func (ls *linuxSystem) DeletePartition(d disko.Disk, number uint) error {
	if err := deletePartitions(d, []uint{number}); err != nil {
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

func (ls *linuxSystem) getDiskType(path string, udInfo disko.UdevInfo) (disko.DiskType, error) {
	ctrl, err := ls.megaraid.Query(0)
	if err == nil {
		for _, vd := range ctrl.VirtDrives {
			if vd.Path == path {
				if ctrl.DriveGroups[vd.DriveGroup].IsSSD() {
					return disko.SSD, nil
				}

				return disko.HDD, nil
			}
		}
	} else if err != megaraid.ErrNoStorcli && err != megaraid.ErrNoController {
		return disko.HDD, err
	}

	return getDiskType(udInfo)
}
