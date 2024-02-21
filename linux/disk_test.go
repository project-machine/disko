package linux

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"machinerun.io/disko"
	"machinerun.io/disko/partid"
)

//nolint:funlen
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
	assert.Equal(
		disko.XENBUS,
		getAttachType(disko.UdevInfo{
			Name:    "xvdf",
			SysPath: "/devices/vbd-51792/block/xvdf",
			Properties: map[string]string{
				"DEVNAME":          "/dev/xvdf",
				"DEVPATH":          "/devices/vbd-51792/block/xvdf",
				"DEVTYPE":          "disk",
				"MAJOR":            "202",
				"MINOR":            "80",
				"SUBSYSTEM":        "block",
				"TAGS":             ":systemd:",
				"USEC_INITIALIZED": "2085116822",
			},
			Symlinks: []string{},
		}))
	assert.Equal(
		disko.NBD,
		getAttachType(disko.UdevInfo{
			Name:    "nbd0",
			SysPath: "/devices/virtual/block/nbd0",
			Properties: map[string]string{
				"DEVNAME":          "/dev/nbd0",
				"DEVPATH":          "/devices/virtual/block/nbd0",
				"DEVTYPE":          "disk",
				"MAJOR":            "43",
				"MINOR":            "0",
				"SUBSYSTEM":        "block",
				"TAGS":             ":systemd:",
				"USEC_INITIALIZED": "781854313926",
			},
			Symlinks: []string{},
		}))
	assert.Equal(
		disko.LOOP,
		getAttachType(disko.UdevInfo{
			Name:    "loop1",
			SysPath: "/devices/virtual/block/loop1",
			Properties: map[string]string{
				"DEVNAME":          "/dev/loop1",
				"DEVPATH":          "/devices/virtual/block/loop1",
				"DEVTYPE":          "disk",
				"MAJOR":            "7",
				"MINOR":            "1",
				"SUBSYSTEM":        "block",
				"TAGS":             ":systemd:",
				"USEC_INITIALIZED": "1264916",
			},
			Symlinks: []string{},
		}))
}

func TestGetDiskType(t *testing.T) {
	assert := assert.New(t)
	systemVirtType = virtKvm
	result, err := getDiskType(disko.UdevInfo{
		Name:    "vdb",
		SysPath: "/devices/pci0000:00/0000:00:03.0/virtio1/block/vdb",
		Properties: map[string]string{
			"DEVLINKS":         "/dev/disk/by-path/pci-0000:00:03.0 /dev/disk/by-path/virtio-pci-0000:00:03.0 /dev/disk/by-id/virtio-ssd-disk0-ssd",
			"DEVNAME":          "/dev/vdb",
			"DEVPATH":          "/devices/pci0000:00/0000:00:03.0/virtio1/block/vdb",
			"DEVTYPE":          "disk",
			"DISKSEQ":          "42",
			"ID_PATH":          "pci-0000:00:03.0",
			"ID_PATH_TAG":      "pci-0000_00_03_0",
			"ID_SERIAL":        "ssd-disk0-ssd",
			"MAJOR":            "253",
			"MINOR":            "16",
			"SUBSYSTEM":        "block",
			"USEC_INITIALIZED": "2422574",
		},
		Symlinks: []string{
			"disk/by-diskseq/42",
			"disk/by-path/pci-0000:00:03.0",
			"disk/by-path/virtio-pci-0000:00:03.0",
			"disk/by-id/virtio-ssd-disk0-ssd",
		},
	})
	if err != nil {
		t.Errorf("expected err = nil, got %s", err)
	}
	assert.Equal(disko.SSD.String(), result.String())
}

func genTempGptDisk(tmpd string, fsize uint64) (disko.Disk, error) {
	fpath := path.Join(tmpd, "mydisk")

	disk := disko.Disk{
		Name:       "mydisk",
		Path:       fpath,
		Size:       fsize,
		SectorSize: sectorSize512,
		Table:      disko.GPT,
	}

	if err := ioutil.WriteFile(fpath, []byte{}, 0600); err != nil {
		return disk, fmt.Errorf("Failed to write to a temp file: %s", err)
	}

	if err := os.Truncate(fpath, int64(fsize)); err != nil {
		return disk, fmt.Errorf("Failed create empty file: %s", err)
	}

	fs := disk.FreeSpaces()
	if len(fs) != 1 {
		return disk, fmt.Errorf("Expected 1 free space, found %d", fs)
	}

	parts := disko.PartitionSet{
		1: disko.Partition{
			Start:  fs[0].Start,
			Last:   fs[0].Last,
			Type:   partid.LinuxLVM,
			Name:   "mytest partition",
			ID:     disko.GenGUID(),
			Number: uint(1),
		}}

	if err := addPartitionSet(disk, parts); err != nil {
		return disk, err
	}

	disk.Partitions = parts

	return disk, nil
}

func TestMyPartition(t *testing.T) {
	tmpd, err := ioutil.TempDir("", "disko_test")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}

	defer os.RemoveAll(tmpd)

	fpath := path.Join(tmpd, "mydisk")
	fsize := uint64(200 * 1024 * 1024)

	if err := ioutil.WriteFile(fpath, []byte{}, 0600); err != nil {
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

	pSet, tType, ssize, err := findPartitions(fp)
	if err != nil {
		t.Errorf("Failed to findPartitions on %s: %s", fpath, err)
	}

	if len(pSet) != 1 {
		t.Errorf("There were %d partitions, expected 1", len(pSet))
	}

	if tType != disko.GPT {
		t.Errorf("Expected GPT partition table, found %s", tType)
	}

	if sectorSize512 != ssize {
		t.Errorf("Expected size %d, found %d", sectorSize512, ssize)
	}

	if pSet[1].ID != myGUID {
		t.Errorf("Guid = %s, not %s", pSet[1].ID.String(), myGUID.String())
	}
}

func TestMyPartitionMBR(t *testing.T) {
	tmpd, err := ioutil.TempDir("", "disko_test")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}

	defer os.RemoveAll(tmpd)

	fpath := path.Join(tmpd, "mydisk")
	fsize := uint64(200 * 1024 * 1024)

	if err := ioutil.WriteFile(fpath, []byte{}, 0600); err != nil {
		t.Fatalf("Failed to write to a temp file: %s", err)
	}

	if err := os.Truncate(fpath, int64(fsize)); err != nil {
		t.Fatalf("Failed create empty file: %s", err)
	}

	disk := disko.Disk{
		Name:       "mydiskMBR",
		Path:       fpath,
		Size:       fsize,
		SectorSize: sectorSize512,
		Table:      disko.MBR,
	}

	fs := disk.FreeSpaces()
	if len(fs) != 1 {
		t.Errorf("Expected 1 free space, found %d", fs)
	}

	part := disko.Partition{
		Start:  fs[0].Start,
		Last:   fs[0].Last,
		Type:   partid.LinuxLVM,
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

	pSet, tType, ssize, err := findPartitions(fp)
	if err != nil {
		t.Errorf("Failed to findPartitions on %s: %s", fpath, err)
	}

	if len(pSet) != 1 {
		t.Errorf("There were %d partitions, expected 1", len(pSet))
	}

	if tType != disko.MBR {
		t.Errorf("Expected GPT partition table, found %s", tType)
	}

	if sectorSize512 != ssize {
		t.Errorf("Expected size %d, found %d", sectorSize512, ssize)
	}
}

//nolint:funlen
func TestWipeDisk(t *testing.T) {
	tmpd, err := ioutil.TempDir("", "disko_test")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}

	mib := uint64(1024 * 1024)

	defer os.RemoveAll(tmpd)

	disk, err := genTempGptDisk(tmpd, 50*mib)
	if err != nil {
		t.Fatalf("Creation of temp disk failed: %s", err)
	}

	if len(disk.Partitions) == 0 {
		t.Fatalf("Found no partitions on the disk from genTempGptDisk")
	}

	fp, err := os.OpenFile(disk.Path, os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("Failed to open disk %s", disk.Path)
	}

	buf := make([]byte, 1024)

	for i := 0; i < 1024; i++ {
		buf[i] = 0xFF
	}

	// write (scribble) 2MiB + 1k of 0xFF at first partition.
	// Wipe should zero the first MiB and other offsets
	if _, err := fp.Seek(int64(disk.Partitions[1].Start), io.SeekStart); err != nil {
		t.Errorf("failed seek to part1 start: %s", err)
	}

	for i := 0; i < (2*int(mib)+1)/len(buf); i++ {
		if n, err := fp.Write(buf); n != len(buf) || err != nil {
			t.Fatalf("failed to write 255 at %d\n", i)
		}
	}
	fp.Close()

	if err := wipeDisk(disk); err != nil {
		t.Errorf("Failed wipe of disk: %s", err)
	}

	fp, err = os.OpenFile(disk.Path, os.O_RDWR, 0)
	if err != nil {
		t.Errorf("Failed opening %s after wipe: %s", disk.Path, err)
	}

	for _, c := range [](struct {
		start uint64
		size  int
		val   byte
		label string
	}){
		{0, int(mib), 0x00, "disk start"},
		{disk.Partitions[1].Start, int(mib), 0x00, "part1 start"},
		{disk.Partitions[1].Start + mib, 4, 0x00, "VMFS vol member got wiped"},
		{disk.Partitions[1].Start + mib + 4, int(mib) - 4, 0xFF, "scribbled 1"},
		{disk.Partitions[1].Start + 2*mib, 4, 0x00, "VMFS got wiped"},
		{disk.Partitions[1].Start + 2*mib + 4, 1, 0x00, "scribble 2"},
		{disk.Size - mib, int(mib), 0x00, "disk end"},
	} {
		if _, err := fp.Seek(int64(c.start), io.SeekStart); err != nil {
			t.Errorf("Failed seek for %s: %s", c.label, err)
			continue
		}

		buf := make([]byte, c.size)
		readlen, err := io.ReadFull(fp, buf)

		if err != nil {
			t.Errorf("Failed read of %d from fp for %s: %s", len(buf), c.label, err)
			continue
		}

		if readlen != c.size {
			t.Errorf("Read %d expected %d for %s: %s", readlen, c.size, c.label, err)
			continue
		}

		for i := 0; i < c.size; i++ {
			if buf[i] != c.val {
				t.Errorf("%s: %d found %d expected %d", c.label, i, buf[i], c.val)
				break
			}
		}
	}

	fp.Close()
}

func TestDeletePartition(t *testing.T) {
	tmpd, err := ioutil.TempDir("", "disko_test")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}

	defer os.RemoveAll(tmpd)

	disk, err := genTempGptDisk(tmpd, 200*1024*1024)
	if err != nil {
		t.Fatalf("Creation of temp disk failed: %s", err)
	}

	fp, err := os.Open(disk.Path)
	if err != nil {
		t.Fatalf("Failed to open file after writing it: %s", err)
	}

	pSet, _, _, err := findPartitions(fp)
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
}

func TestUpdatePartition(t *testing.T) {
	tmpd, err := ioutil.TempDir("", "disko_test")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}

	defer os.RemoveAll(tmpd)

	disk, err := genTempGptDisk(tmpd, 200*1024*1024)
	if err != nil {
		t.Fatalf("Creation of temp disk failed: %s", err)
	}

	fp, err := os.Open(disk.Path)
	if err != nil {
		t.Fatalf("Failed to open file after writing it: %s", err)
	}

	pSet, _, _, err := findPartitions(fp)
	if err != nil {
		t.Fatalf("Failed to find partitions")
	}

	pOrig := pSet[1]
	numOrig := len(pSet)

	fp.Close()

	myID, err := disko.StringToGUID("ABCDEF01-2345-6789-ABCD-EF0123456789")
	if err != nil {
		t.Fatalf("Failed to convert guid")
	}

	myType, err := disko.StringToPartType("A0B0C0D0-E0F0-A1B1-C1D1-E1F1A2B2C2D2")
	if err != nil {
		t.Fatalf("Failed to convert string to parttype")
	}

	newPSet := disko.PartitionSet{
		1: disko.Partition{
			Type:   myType,
			Name:   "new name",
			ID:     myID,
			Number: 1,
		},
	}

	err = updatePartitions(disk, newPSet)
	if err != nil {
		t.Fatalf("Failed update partition 1: %s", err)
	}

	fp, err = os.Open(disk.Path)
	if err != nil {
		t.Fatalf("Failed to open file after writing it: %s", err)
	}

	pSet, _, _, err = findPartitions(fp)
	if err != nil {
		t.Fatalf("Failed to re-findPartitions on %s: %s", disk.Path, err)
	}

	fp.Close()

	expected := disko.Partition{
		Type:   myType,
		Name:   "new name",
		ID:     myID,
		Number: 1,
		Start:  pOrig.Start,
		Last:   pOrig.Last,
	}
	assert := assert.New(t)
	assert.Equal(expected, pSet[1])

	assert.Equalf(numOrig, len(pSet), "Expected %d partitions, but now have %d", numOrig, len(pSet))
}

func TestBadPartition(t *testing.T) {
	tmpd, err := ioutil.TempDir("", "disko_test")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}

	defer os.RemoveAll(tmpd)

	fpath := path.Join(tmpd, "mydisk")
	fsize := uint64(200 * 1024 * 1024)

	if err := ioutil.WriteFile(fpath, []byte{}, 0600); err != nil {
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
	myGUID := disko.GenGUID()

	part := disko.Partition{
		Start:  1024,
		Last:   fs[0].Last,
		Type:   partid.LinuxLVM,
		Name:   "mytest partition",
		ID:     myGUID,
		Number: uint(1),
	}

	err = addPartitionSet(disk, disko.PartitionSet{part.Number: part})
	if err == nil {
		t.Errorf("Created partition with OOB start (%d). should have failed", part.Start)
	}

	part.Start = fs[0].Start
	part.Last = disk.Size - 1

	err = addPartitionSet(disk, disko.PartitionSet{part.Number: part})
	if err == nil {
		t.Errorf("Created partition with OOB end (%d). should have failed", part.Last)
	}
}

func TestEmptyPartIDIsFilled(t *testing.T) {
	empty := disko.GUID{
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	}
	before := disko.Partition{
		Start:  4 * 1024 * 1024,
		Last:   8*1024*1024 - 1,
		Type:   partid.LinuxLVM,
		ID:     empty,
		Number: uint(1),
	}

	after := toGPTPartition(before, sectorSize512)
	if before.ID == disko.GUID(after.Id) {
		t.Errorf("Empty ID did not get filled in toGPTPartition.")
	}

	after2 := toGPTPartition(before, sectorSize512)
	if after2.Id == after.Id {
		t.Errorf("toGPTPartition returned 2 identical ID GUIDs: %s.", after.Id)
	}
}

func TestNoEmptyPartID(t *testing.T) {
	myID := disko.GUID{
		0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8,
		0x9, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16,
	}
	before := disko.Partition{
		Start:  4 * 1024 * 1024,
		Last:   8*1024*1024 - 1,
		Type:   partid.LinuxLVM,
		ID:     myID,
		Number: uint(1),
	}

	after := toGPTPartition(before, sectorSize512)
	if before.ID != disko.GUID(after.Id) {
		t.Errorf("toGPTPartition changed partition ID: %s -> %s", before.ID, after.Id)
	}
}
