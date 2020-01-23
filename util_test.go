package disko

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindGaps0(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{{0, 100}},
		findRangeGaps([]uRange{}, 0, 100))
}

func TestFindGaps1(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{{0, 49}, {60, 100}},
		findRangeGaps([]uRange{{50, 59}}, 0, 100))
}

func TestFindGaps2(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{{51, 100}},
		findRangeGaps([]uRange{{0, 50}}, 0, 100))
}

func TestFindGaps3(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{{0, 10}},
		findRangeGaps([]uRange{{11, 100}}, 0, 100))
}

func TestFindGaps4(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{{0, 10}, {50, 59}, {91, 100}},
		findRangeGaps([]uRange{{11, 49}, {60, 90}}, 0, 100))
}

func TestFindGaps5(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{},
		findRangeGaps([]uRange{{0, 10}, {11, 100}}, 0, 100))
}

func TestFindGaps6(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{},
		findRangeGaps([]uRange{{0, 150}, {50, 100}}, 0, 100))
}

func TestFindGaps7(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{{0, 9}, {41, 49}, {101, 110}},
		findRangeGaps([]uRange{{10, 40}, {50, 100}}, 0, 110))
}

func TestFindGaps8(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]uRange{{10, 100}},
		findRangeGaps([]uRange{{110, 200}}, 10, 100))
}

func TestParseUdevInfo(t *testing.T) {
	data := []byte(`P: /devices/virtual/block/dm-0
N: dm-0
S: disk/by-id/dm-name-nvme0n1p6_crypt
S: disk/by-id/dm-uuid-CRYPT-LUKS1-b174c64e7a714359a8b56b79fb66e92b-nvme0n1p6_crypt
S: disk/by-uuid/25df9069-80c7-46f4-a47c-305613c2cb6b
S: mapper/nvme0n1p6_crypt
E: DEVLINKS=/dev/disk/by-id/dm-uuid-CRYPT-LUKS1-b174b-nvme0n1p6_crypt ` +
		`/dev/mapper/nvme0n1p6_crypt /dev/disk/by-id/dm-name-nvme0n1p6_crypt
E: DEVNAME=/dev/dm-0
`)
	myInfo := UdevInfo{}
	ast := assert.New(t)

	err := parseUdevInfo(data, &myInfo)
	ast.Equal(err, nil)
	ast.Equal(
		UdevInfo{
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
	myInfo := UdevInfo{}
	err := parseUdevInfo(data, &myInfo)
	ast.Equal(nil, err)
	ast.Equal(
		UdevInfo{
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
