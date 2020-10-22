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

func TestGetDiskProperties(t *testing.T) {
	azureSys := ("/devices/LNXSYSTM:00/LNXSYBUS:00/PNP0A03:00/device:07/VMBUS:01" +
		"/00000000-0001-8899-0000-000000000000/host1/target1:0:1/1:0:1:0/block/sdb")
	scsiSys := "/devices/pci0000:00/0000:00:02.2/0000:05:00.0/host0/target0:0:8/0:0:8:0/block/sda"

	tables := []struct {
		info     disko.UdevInfo
		expected disko.PropertySet
	}{
		{
			disko.UdevInfo{
				Name:     "sda",
				SysPath:  scsiSys,
				Symlinks: []string{},
				Properties: map[string]string{
					"ID_MODEL":    "SPCC M.2 PCIe SSD",
					"ID_REVISION": "ECFM22.6"}},
			disko.PropertySet{disko.Ephemeral: false}},
		{
			disko.UdevInfo{
				Name:     "sdb",
				SysPath:  azureSys,
				Symlinks: []string{},
				Properties: map[string]string{
					"ID_MODEL":    "SPCC M.2 PCIe SSD",
					"ID_REVISION": "ECFM22.6"}},
			disko.PropertySet{disko.Ephemeral: true}},
		{
			disko.UdevInfo{
				Name:     "sdb",
				SysPath:  azureSys,
				Symlinks: []string{},
				Properties: map[string]string{
					"DM_MULTIPATH_DEVICE_PATH": "0",
					"ID_SERIAL_SHORT":          "AWS628703BD8E5BEB551",
					"ID_WWN":                   "nvme.1d0f-4157...4616e63652053746f72616765-00000001",
					"ID_MODEL":                 "Amazon EC2 NVMe Instance Storage",
					"ID_REVISION":              "0",
					"ID_SERIAL":                "Amazon EC2 NVMe Instance Storage_AWS628703BD8E5BEB551"}},

			disko.PropertySet{disko.Ephemeral: true}},
	}

	for _, table := range tables {
		found := getDiskProperties(table.info)
		bad := []disko.Property{}

		for k, v := range table.expected {
			if found[k] != v {
				bad = append(bad, k)
			}
		}

		for k, v := range found {
			if table.expected[k] != v {
				bad = append(bad, k)
			}
		}

		if len(bad) != 0 {
			t.Errorf("getDiskProperties(%v) returned '%v'. expected '%v'",
				table.info, found, table.expected)
		}
	}
}

func genEmptyDisk(tmpd string, fsize uint64) (disko.Disk, error) {
	fpath := path.Join(tmpd, "mydisk")

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

	return disk, nil
}

func TestCreatePartitions(t *testing.T) {
	ast := assert.New(t)

	tmpd, err := ioutil.TempDir("", "disko_test")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}

	defer os.RemoveAll(tmpd)

	disk, err := genEmptyDisk(tmpd, 50*disko.Mebibyte)
	if err != nil {
		t.Fatalf("Creation of temp disk failed: %s", err)
	}

	part1 := disko.Partition{
		Start:  4 * disko.Mebibyte,
		Last:   20*disko.Mebibyte - 1,
		Type:   partid.LinuxHome,
		Name:   "mytest 1",
		ID:     disko.GenGUID(),
		Number: uint(1),
	}

	part2 := disko.Partition{
		Start:  20 * disko.Mebibyte,
		Last:   40*disko.Mebibyte - 1,
		Type:   partid.LinuxFS,
		Name:   "mytest 2",
		ID:     disko.GenGUID(),
		Number: uint(2),
	}

	pSet := disko.PartitionSet{1: part1, 2: part2}

	sys := System()
	if err := sys.CreatePartitions(disk, pSet); err != nil {
		t.Errorf("CreatePartitions failed: %s", err)
	}

	fp, err := os.Open(disk.Path)
	if err != nil {
		t.Fatalf("Failed to open disk image %s: %s", disk.Path, err)
	}

	pSetFound, _, _, err := findPartitions(fp)
	if err != nil {
		t.Fatalf("Failed to findPartitions on %s: %s", disk.Path, err)
	}

	if len(pSetFound) != len(pSet) {
		t.Errorf("Scanned found %d partitions, expected %d", len(pSetFound), len(pSet))
	}

	ast.Equal(part1, pSetFound[1])
	ast.Equal(part2, pSetFound[2])
}
