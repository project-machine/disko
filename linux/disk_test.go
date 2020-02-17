// +build linux

package linux

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/anuvu/disko"
	"github.com/anuvu/disko/partid"
	"github.com/stretchr/testify/assert"
)

func TestGetAttachType(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		disko.VIRTIO,
		getAttachType(disko.UdevInfo{
			Name:       "vda",
			SysPath:    "/devices/pci0000:00/0000:00:05.0/virtio3/block/vda",
			Properties: map[string]string{},
			Symlinks:   []string{"disk/by-path/pci-0000:00:05.0"},
		}))

	assert.Equal(
		disko.ATA,
		getAttachType(disko.UdevInfo{
			Name:    "sda",
			SysPath: "/devices/pci0000:00/0000:00:0d.0/host0/target0:0:0/0:0:0:0/block/sda",
			Properties: map[string]string{
				"ID_BUS": "ata",
			},
			Symlinks: []string{"disk/by-id/ata-VBOX_HARDDISK_VB579a85b0-bf6debae"},
		}))
	assert.Equal(
		disko.USB,
		getAttachType(disko.UdevInfo{
			Name:    "sda",
			SysPath: "/devices/pci0000:00/0000:00:14.0/usb2/2-3/2-3:1.0/host0/target0:0:0/0:0:0:0/block/sda",
			Properties: map[string]string{
				"ID_BUS": "usb",
			},
			Symlinks: []string{"disk/by-id/ata-VBOX_HARDDISK_VB579a85b0-bf6debae"},
		}))
	assert.Equal(
		disko.SCSI,
		getAttachType(disko.UdevInfo{
			Name:    "sda",
			SysPath: "/devices/pci0000:00/0000:00:02.2/0000:05:00.0/host0/target0:0:8/0:0:8:0/block/sda",
			Properties: map[string]string{
				"ID_BUS": "scsi",
			},
			Symlinks: []string{"disk/by-id/scsi-35000c500a0d8963f",
				"disk/by-id/wwn-0x5000c500a0d8963f"},
		}))
	assert.Equal(
		disko.PCIE,
		getAttachType(disko.UdevInfo{
			Name:       "nvme0p1",
			SysPath:    "/devices/pci0000:00/0000:00:1c.4/0000:04:00.0/nvme/nvme0/nvme0n1",
			Properties: map[string]string{},
			Symlinks: []string{"disk/by-id/nvme-SPCC_M.2_PCIe_SSD_BD52079C067D00486555",
				"disk/by-id/nvme-eui.6479a72be0043535"},
		}))
}

func genTempGptDisk(tmpd string) (disko.Disk, error) {
	fpath := path.Join(tmpd, "mydisk")
	fsize := uint64(200 * 1024 * 1024) // nolint:gomnd (200MiB)

	disk := disko.Disk{
		Name:       "mydisk",
		Path:       fpath,
		Size:       fsize,
		SectorSize: sectorSize512,
	}

	if err := ioutil.WriteFile(fpath, []byte{}, 0644); err != nil {
		return disk, fmt.Errorf("Failed to write to a temp file: %s", err)
	}

	if err := os.Truncate(fpath, int64(fsize)); err != nil {
		return disk, fmt.Errorf("Failed create empty file: %s", err)
	}

	fs := disk.FreeSpaces()
	if len(fs) != 1 {
		return disk, fmt.Errorf("Expected 1 free space, found %d", fs)
	}

	part := disko.Partition{
		Start:  fs[0].Start,
		Last:   fs[0].Last,
		Type:   partid.LinuxLVM,
		Name:   "mytest partition",
		ID:     disko.GenGUID(),
		Number: uint(1),
	}

	if err := addPartitionSet(disk, disko.PartitionSet{part.Number: part}); err != nil {
		return disk, err
	}

	return disk, nil
}

func TestMyPartition(t *testing.T) {
	tmpd, err := ioutil.TempDir("", "disko_test")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}

	defer os.RemoveAll(tmpd)

	fpath := path.Join(tmpd, "mydisk")
	fsize := uint64(200 * 1024 * 1024) // nolint:gomnd (200MiB)

	if err := ioutil.WriteFile(fpath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to write to a temp file: %s", err)
	}

	if err := os.Truncate(fpath, int64(fsize)); err != nil {
		t.Fatalf("Failed create empty file: %s", err)
	}

	disk := disko.Disk{
		Name:       "mydisk",
		Path:       fpath,
		Size:       fsize,
		SectorSize: sectorSize512,
	}

	fs := disk.FreeSpaces()
	if len(fs) != 1 {
		t.Errorf("Expected 1 free space, found %d", fs)
	}

	myGUID := disko.GenGUID()

	part := disko.Partition{
		Start:  fs[0].Start,
		Last:   fs[0].Last,
		Type:   partid.LinuxLVM,
		Name:   "mytest partition",
		ID:     myGUID,
		Number: uint(1),
	}

	err = addPartitionSet(disk, disko.PartitionSet{part.Number: part})
	if err != nil {
		t.Errorf("Creation of partition failed: %s", err)
	}

	fp, err := os.Open(fpath)
	if err != nil {
		t.Fatalf("Failed to open file after writing it: %s", err)
	}

	pSet, ssize, err := findPartitions(fp)
	if err != nil {
		t.Errorf("Failed to findPartitions on %s: %s", fpath, err)
	}

	if len(pSet) != 1 {
		t.Errorf("There were %d partitions, expected 1", len(pSet))
	}

	if sectorSize512 != ssize {
		t.Errorf("Expected size %d, found %d", sectorSize512, ssize)
	}

	if pSet[1].ID != myGUID {
		t.Errorf("Guid = %s, not %s", pSet[1].ID.String(), myGUID.String())
	}
}

func TestDeletePartition(t *testing.T) {
	tmpd, err := ioutil.TempDir("", "disko_test")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}

	defer os.RemoveAll(tmpd)

	disk, err := genTempGptDisk(tmpd)
	if err != nil {
		t.Fatalf("Creation of temp disk failed: %s", err)
	}

	fp, err := os.Open(disk.Path)
	if err != nil {
		t.Fatalf("Failed to open file after writing it: %s", err)
	}

	pSet, _, err := findPartitions(fp)
	if err != nil {
		t.Fatalf("Failed to findPartitions on %s: %s", disk.Path, err)
	}

	if len(pSet) != 1 {
		t.Fatalf("There were %d partitions, expected 1", len(pSet))
	}

	err = deletePartitions(disk, []uint{1})
	if err != nil {
		t.Fatalf("Failed delete partition 1: %s", err)
	}

	pSet, _, err = findPartitions(fp)
	if err != nil {
		t.Fatalf("Failed to re-findPartitions on %s: %s", disk.Path, err)
	}

	if len(pSet) != 0 {
		t.Fatalf("There were %d partitions after delete, expected 0", len(pSet))
	}
}
