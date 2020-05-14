package disko_test

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/anuvu/disko"
	"github.com/anuvu/disko/partid"
)

func TestFreeSpaceSize(t *testing.T) {
	values := []struct{ start, last, expected uint64 }{
		{0, 9, 10},
		{0, 199, 200},
		{100, 200, 101},
	}

	for _, v := range values {
		f := disko.FreeSpace{v.start, v.last}
		found := f.Size()

		if v.expected != found {
			t.Errorf("Size(%v) expected %d found %d",
				f, v.expected, found)
		}
	}
}

func TestPartitionSize(t *testing.T) {
	tables := []struct{ start, last, expected uint64 }{
		{0, 99, 100},
		{3 * disko.Mebibyte, (5000+3)*disko.Mebibyte - 1, 5000 * disko.Mebibyte},
	}

	for _, table := range tables {
		p := disko.Partition{Start: table.start, Last: table.last}
		found := p.Size()

		if table.expected != found {
			t.Errorf("Size(%v) expected %d found %d", p, table.expected, found)
		}
	}
}

func TestDiskString(t *testing.T) {
	mib := disko.Mebibyte
	gb := uint64(1000 * 1000 * 1000) // nolint: gomnd

	d := disko.Disk{
		Name:       "sde",
		Path:       "/dev/sde",
		Size:       gb,
		SectorSize: 512, //nolint: gomnd
		Type:       disko.HDD,
		Attachment: disko.ATA,
		Partitions: disko.PartitionSet{
			1: {Start: 3 * mib, Last: 253*mib - 1, Number: 1},   //nolint: gomnd
			3: {Start: 500 * mib, Last: 600*mib - 1, Number: 3}, //nolint: gomnd
		},
		UdevInfo: disko.UdevInfo{},
	}
	found := " " + d.String() + " "

	// disk size 1gb = 953 MiB. 600 = (253-3) + (953-600)
	expectedFree := 600 // nolint: gomnd

	for _, substr := range []string{
		fmt.Sprintf("Size=%d", gb),
		fmt.Sprintf("FreeSpace=%dMiB/2", expectedFree), //nolint: gomnd
		fmt.Sprintf("NumParts=%d", len(d.Partitions))} {
		if !strings.Contains(found, " "+substr+" ") {
			t.Errorf("%s: missing expected substring ' %s '", found, substr)
		}
	}
}

func TestDiskDetails(t *testing.T) {
	mib := disko.Mebibyte
	d := disko.Disk{
		Name:       "sde",
		Path:       "/dev/sde",
		Size:       mib * mib,
		SectorSize: 512, //nolint: gomnd
		Type:       disko.HDD,
		Attachment: disko.ATA,
		Partitions: disko.PartitionSet{
			1: {Start: 3 * mib, Last: 253*mib - 1, Number: 1}, //nolint: gomnd
		},
		UdevInfo: disko.UdevInfo{},
	}
	expected := `
[ # Start Last Size Name ]
[ 1 3 MiB 253 MiB 250 MiB                 ]
[ - 253 MiB 1048575 MiB 1048322 MiB <free> ]`

	spaces := regexp.MustCompile("[ ]+")
	found := strings.TrimSpace(spaces.ReplaceAllString(d.Details(), " "))
	expShort := strings.TrimSpace(spaces.ReplaceAllString(expected, " "))

	if expShort != found {
		t.Errorf("Expected: '%s'\nFound: '%s'\n", expShort, found)
	}
}

func TestDiskTypeString(t *testing.T) {
	for _, d := range []struct {
		dtype    disko.DiskType
		expected string
	}{
		{disko.HDD, "HDD"},
		{disko.SSD, "SSD"},
		{disko.NVME, "NVME"},
	} {
		found := d.dtype.String()
		if found != d.expected {
			t.Errorf("disko.DiskType(%d).String() found %s, expected %s",
				d.dtype, found, d.expected)
		}
	}
}

func TestAttachmentTypeString(t *testing.T) {
	for _, d := range []struct {
		dtype    disko.AttachmentType
		expected string
	}{
		{disko.UnknownAttach, "UNKNOWN"},
		{disko.RAID, "RAID"},
		{disko.SCSI, "SCSI"},
		{disko.ATA, "ATA"},
		{disko.PCIE, "PCIE"},
		{disko.USB, "USB"},
		{disko.VIRTIO, "VIRTIO"},
		{disko.IDE, "IDE"},
		{disko.NBD, "NBD"},
		{disko.LOOP, "LOOP"},
	} {
		found := d.dtype.String()
		if found != d.expected {
			t.Errorf("disko.AttachmentType(%d).String() found %s, expected %s",
				d.dtype, found, d.expected)
		}
	}
}

func TestPartitionSerializeJson(t *testing.T) {
	// For readability, Partition serializes ID and Type to string GUIDs
	// Test that they get there.
	myIDStr := "01234567-89AB-CDEF-0123-456789ABCDEF"
	myID, _ := disko.StringToGUID(myIDStr)
	p := disko.Partition{
		Start:  3 * disko.Mebibyte, //nolint:gomnd
		Last:   253*disko.Mebibyte - 1,
		ID:     myID,
		Type:   partid.EFI,
		Name:   "my system part",
		Number: 1,
	}

	jbytes, err := json.MarshalIndent(&p, "", "  ")
	if err != nil {
		t.Errorf("Failed to marshal %#v: %s", p, err)
	}

	jstr := string(jbytes)
	if !strings.Contains(jstr, myIDStr) {
		t.Errorf("Did not find string ID '%s' in json: %s", myIDStr, jstr)
	}

	typeStr := disko.GUIDToString(disko.GUID(partid.EFI))
	if !strings.Contains(jstr, typeStr) {
		t.Errorf("Did not find string Type '%s' in json: %s", myIDStr, jstr)
	}

	fmt.Printf("%s\n", jstr)
}

func TestPartitionUnserializeJson(t *testing.T) {
	myIDStr := "01234567-89AB-CDEF-0123-456789ABCDEF"
	myID, _ := disko.StringToGUID(myIDStr)
	jbytes := []byte(`{
  "start": 3145728,
  "last": 265289727,
  "id": "01234567-89AB-CDEF-0123-456789ABCDEF",
  "type": "C12A7328-F81F-11D2-BA4B-00A0C93EC93B",
  "name": "my system part",
  "number": 1
}`)

	found := disko.Partition{}

	err := json.Unmarshal(jbytes, &found)
	if err != nil {
		t.Errorf("Failed Unmarshal of bytes to Partition: %s", err)
	}

	expected := disko.Partition{
		Start:  3 * disko.Mebibyte,     // nolint:gomnd
		Last:   253*disko.Mebibyte - 1, // nolint:gomnd
		ID:     myID,
		Type:   partid.EFI,
		Name:   "my system part",
		Number: 1,
	}

	if expected != found {
		t.Errorf("Objects differed. got %#v expected %#v\n", found, expected)
	}
}

func TestDiskSerializeJson(t *testing.T) {
	// For readability, Partition serializes ID and Type to string GUIDs
	// Test that they get there.
	d := disko.Disk{
		Name:       "sda",
		Path:       "/dev/sda",
		Size:       500 * disko.Mebibyte, //nolint:gomnd
		SectorSize: 512,                  //nolint:gomnd
		Type:       disko.HDD,
		Attachment: disko.ATA,
	}

	jbytes, err := json.MarshalIndent(&d, "", "  ")
	if err != nil {
		t.Errorf("Failed to marshal %#v: %s", d, err)
	}

	jstr := string(jbytes)
	if !strings.Contains(jstr, "HDD") {
		t.Errorf("Did not find string 'HDD' in json: %s", jstr)
	}
}

func compareDisk(a *disko.Disk, b *disko.Disk) bool {
	return (a.Name == b.Name &&
		a.Path == b.Path &&
		a.Size == b.Size &&
		a.SectorSize == b.SectorSize &&
		a.Type == b.Type &&
		a.Attachment == b.Attachment)
}

func TestDiskUnserializeJson(t *testing.T) {
	expected := disko.Disk{
		Name:       "sda",
		Path:       "/dev/sda",
		Size:       500 * disko.Mebibyte, //nolint:gomnd
		SectorSize: 512,                  //nolint:gomnd
		Type:       disko.HDD,
		Attachment: disko.ATA,
	}

	for _, jbytes := range [][]byte{
		[]byte(`{
  "name": "sda",
  "path": "/dev/sda",
  "size": 524288000,
  "sectorSize": 512,
  "type": "HDD",
  "attachment": "ATA"}`),
		[]byte(`{
  "name": "sda",
  "path": "/dev/sda",
  "size": 524288000,
  "sectorSize": 512,
  "type": 0,
  "attachment": 3}`)} {
		found := disko.Disk{}
		err := json.Unmarshal(jbytes, &found)

		if err != nil {
			t.Errorf("Failed Unmarshal of bytes to Disk: %s", err)
		}

		if !compareDisk(&found, &expected) {
			t.Errorf("Objects differed. got %#v expected %#v\n", found, expected)
		}
	}
}
