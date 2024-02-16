package linux

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"machinerun.io/disko"
)

func TestParseUdevInfo(t *testing.T) {
	data := []byte(`P: /devices/virtual/block/dm-0
N: dm-0
M: dm-0
S: disk/by-id/dm-name-nvme0n1p6_crypt
S: disk/by-id/dm-uuid-CRYPT-LUKS1-b174c64e7a714359a8b56b79fb66e92b-nvme0n1p6_crypt
S: disk/by-uuid/25df9069-80c7-46f4-a47c-305613c2cb6b
S: mapper/nvme0n1p6_crypt
E: DEVLINKS=/dev/disk/by-id/dm-uuid-CRYPT-LUKS1-b174b-nvme0n1p6_crypt ` +
		`/dev/mapper/nvme0n1p6_crypt /dev/disk/by-id/dm-name-nvme0n1p6_crypt
E: DEVNAME=/dev/dm-0
`)

	ast := assert.New(t)

	myInfo := disko.UdevInfo{}
	ast.Nil(parseUdevInfo(data, &myInfo))

	ast.Equal(
		disko.UdevInfo{
			Name:    "dm-0",
			SysPath: "/devices/virtual/block/dm-0",
			Symlinks: []string{
				"disk/by-id/dm-name-nvme0n1p6_crypt",

				"disk/by-id/dm-uuid-CRYPT-LUKS1-b174c64e7a714359a8b56b79fb66e92b-nvme0n1p6_crypt",
				"disk/by-uuid/25df9069-80c7-46f4-a47c-305613c2cb6b",
				"mapper/nvme0n1p6_crypt",
			},
			Properties: map[string]string{
				"DEVLINKS": ("/dev/disk/by-id/dm-uuid-CRYPT-LUKS1-b174b-nvme0n1p6_crypt /dev/mapper/nvme0n1p6_crypt " +
					"/dev/disk/by-id/dm-name-nvme0n1p6_crypt"),
				"DEVNAME": "/dev/dm-0",
			},
		},
		myInfo)
}

func TestParseUdevInfo2(t *testing.T) {
	data := []byte(`P: /devices/pci0000:00/..../block/sda
N: sda
M: sda
S: disk/by-id/scsi-35000c500a0d8963f
S: disk/by-id/wwn-0x5000c500a0d8963f
S: disk/by-path/pci-0000:05:00.0-scsi-0:0:8:0
E: DEVLINKS=/dev/disk/by-path/pci-0000:05:00.0-scsi-0:0:8:0
E: DEVNAME=/dev/sda
E: DEVTYPE=disk
E: ID_BUS=scsi
E: ID_MODEL=ST1000NX0453
E: ID_MODEL_ENC=ST\x2f1000NX0453\x20\x20\x20
E: ID_VENDOR_ENC=SEAGATE\x20
E: ID_WWN=0x5000c500a0d8963f
E: MAJOR=8
E: MINOR=0
E: SUBSYSTEM=block
E: TAGS=:systemd:
E: USEC_INITIALIZED=1926114
`)
	ast := assert.New(t)
	myInfo := disko.UdevInfo{}
	err := parseUdevInfo(data, &myInfo)
	ast.Equal(nil, err)
	ast.Equal(
		disko.UdevInfo{
			Name:    "sda",
			SysPath: "/devices/pci0000:00/..../block/sda",
			Symlinks: []string{
				"disk/by-id/scsi-35000c500a0d8963f",

				"disk/by-id/wwn-0x5000c500a0d8963f",
				"disk/by-path/pci-0000:05:00.0-scsi-0:0:8:0",
			},
			Properties: map[string]string{
				"DEVLINKS":         "/dev/disk/by-path/pci-0000:05:00.0-scsi-0:0:8:0",
				"DEVNAME":          "/dev/sda",
				"DEVTYPE":          "disk",
				"ID_BUS":           "scsi",
				"ID_MODEL":         "ST1000NX0453",
				"ID_MODEL_ENC":     "ST/1000NX0453",
				"ID_VENDOR_ENC":    "SEAGATE",
				"ID_WWN":           "0x5000c500a0d8963f",
				"MAJOR":            "8",
				"MINOR":            "0",
				"SUBSYSTEM":        "block",
				"TAGS":             ":systemd:",
				"USEC_INITIALIZED": "1926114",
			},
		},
		myInfo)
}

func TestRunCommandWithOutputErrorRc(t *testing.T) {
	assert := assert.New(t)
	out, err, rc := runCommandWithOutputErrorRc(
		"sh", "-c", "echo -n STDOUT; echo STDERR 1>&2; exit 99")
	assert.Equal(out, []byte("STDOUT"))
	assert.Equal(err, []byte("STDERR\n"))
	assert.Equal(rc, 99)
}

func TestRunCommandWithOutputErrorRcStdin(t *testing.T) {
	assert := assert.New(t)
	out, err, rc := runCommandWithOutputErrorRcStdin(
		"line1\nline2\n0\n",
		"sh", "-c",
		`read o; echo "$o"; read o; echo "$o" 1>&2; read rc; exit $rc`)
	assert.Equal(out, []byte("line1\n"))
	assert.Equal(err, []byte("line2\n"))
	assert.Equal(rc, 0)
}

func TestRunCommandWithStdin(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(runCommandStdin("the-stdin", "sh", "-c", "exit 0"))
	assert.NotNil(runCommandStdin("", "sh", "-c", "exit 1"))
}

func TestRunCommand(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(runCommand("sh", "-c", "exit 0"))
	assert.NotNil(runCommand("sh", "-c", "exit 1"))
}

func TestCeilingUp(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(uint64(100), Ceiling(98, 4))
}

func TestCeilingEven(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(uint64(100), Ceiling(100, 4))
	assert.Equal(uint64(97), Ceiling(97, 1))
}

func TestFloorDown(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(uint64(96), Floor(98, 4))
}

func TestFloorEven(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(uint64(100), Floor(100, 4))
	assert.Equal(uint64(97), Floor(97, 1))
}

func TestGetFileSize(t *testing.T) {
	data := "This is my data in the file"

	fp, err := ioutil.TempFile("", "testSize")
	defer os.Remove(fp.Name())

	if err != nil {
		t.Fatalf("Failed to make test file: %s", err)
	}

	if _, err := fp.WriteString(data); err != nil {
		t.Fatalf("failed writing to file %s: %s", fp.Name(), err)
	}

	if err := fp.Sync(); err != nil {
		t.Fatal("failed sync")
	}

	found, err := getFileSize(fp)
	if err != nil {
		t.Errorf("Failed to getFileSize: %s", err)
	}

	if found != uint64(len(data)) {
		t.Errorf("Found size %d expected %d", found, len(data))
	}
}

func TestLvPath(t *testing.T) {
	tables := []struct{ vgName, lvName, expected string }{
		{"vg0", "lv0", "/dev/vg0/lv0"},
		{"vg0", "my-foo_bar", "/dev/vg0/my-foo_bar"},
	}

	for _, table := range tables {
		found := lvPath(table.vgName, table.lvName)
		if found != table.expected {
			t.Errorf("lvPath(%s, %s) returned '%s'. expected '%s'",
				table.vgName, table.lvName, found, table.expected)
		}
	}
}

func TestVgLv(t *testing.T) {
	tables := []struct{ vgName, lvName, expected string }{
		{"vg0", "lv0", "vg0/lv0"},
		{"vg0", "my-foo_bar", "vg0/my-foo_bar"},
	}

	for _, table := range tables {
		found := vgLv(table.vgName, table.lvName)
		if found != table.expected {
			t.Errorf("vgLv(%s, %s) returned '%s'. expected '%s'",
				table.vgName, table.lvName, found, table.expected)
		}
	}
}
