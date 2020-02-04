package disko_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/anuvu/disko"
)

func TestFreeSpaceSize(t *testing.T) {
	values := []struct{ start, end, expected uint64 }{
		{0, 9, 10},
		{0, 199, 200},
		{100, 200, 101},
	}

	for _, v := range values {
		f := disko.FreeSpace{v.start, v.end}
		found := f.Size()

		if v.expected != found {
			t.Errorf("Size(%v) expected %d found %d",
				f, v.expected, found)
		}
	}
}

func TestPartitionSize(t *testing.T) {
	tables := []struct{ start, end, expected uint64 }{
		{0, 99, 100},
		{3 * disko.Mebibyte, (5000+3)*disko.Mebibyte - 1, 5000 * disko.Mebibyte},
	}

	for _, table := range tables {
		p := disko.Partition{Start: table.start, End: table.end}
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
			1: {Start: 3 * mib, End: 253*mib - 1, Number: 1},   //nolint: gomnd
			3: {Start: 500 * mib, End: 600*mib - 1, Number: 3}, //nolint: gomnd
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
			1: {Start: 3 * mib, End: 253*mib - 1, Number: 1}, //nolint: gomnd
		},
		UdevInfo: disko.UdevInfo{},
	}
	expected := `
[ # Start End Size Name ]
[ 1 3 MiB 253 MiB 250 MiB                 ]
[ - 253 MiB 1048575 MiB 1048322 MiB <free> ]`

	spaces := regexp.MustCompile("[ ]+")
	found := strings.TrimSpace(spaces.ReplaceAllString(d.Details(), " "))
	expShort := strings.TrimSpace(spaces.ReplaceAllString(expected, " "))

	if expShort != found {
		t.Errorf("Expected: '%s'\nFound: '%s'\n", expShort, found)
	}
}
