// +build linux,!skipIntegration

// nolint:errcheck,gomnd
package linux_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/anuvu/disko"
	"github.com/anuvu/disko/linux"
	"github.com/anuvu/disko/partid"
	"github.com/stretchr/testify/assert"
)

const MiB = 1024 * 1024
const GiB = MiB * 1024

// runLog - run command and Printf, useful for debugging errors.
func runLog(args ...string) {
	out, err, rc := runCommandWithOutputErrorRc(args...)
	fmt.Printf("%s\n", cmdString(args, out, err, rc))
}

func TestPartition(t *testing.T) {
	skipIfNoLoop(t)
	var loopDev string

	ast := assert.New(t)
	tmpFile := getTempFile(GiB)

	defer os.Remove(tmpFile)

	if cleanup, path, err := connectLoop(tmpFile); err != nil {
		runLog("losetup", "-a")
		t.Fatalf("failed loop: %s\n", err)
	} else {
		defer cleanup()
		loopDev = path
	}

	lSys := linux.System()

	disk, err := lSys.ScanDisk(loopDev)
	if err != nil {
		t.Fatalf("Failed first scan of %s: %s\n", loopDev, err)
	}

	part1 := disko.Partition{
		Start:  disk.FreeSpaces()[0].Start,
		Last:   disk.FreeSpaces()[0].Start + (100 * MiB) - 1,
		Type:   disko.PartType(partid.LinuxRootX86),
		Name:   randStr(10),
		Number: 1,
	}

	// part3 leaves 100MiB gap to verify FreeSpaces
	part3 := disko.Partition{
		Start:  part1.Last + 100*MiB + 1,
		Last:   disk.FreeSpaces()[0].Last,
		Type:   disko.PartType(partid.LinuxFS),
		Name:   randStr(8),
		Number: 3,
	}

	if err = lSys.CreatePartition(disk, part1); err != nil {
		t.Fatalf("failed create partition %#v", part1)
	}

	if err = lSys.CreatePartition(disk, part3); err != nil {
		t.Fatalf("failed create partition %#v", part3)
	}

	disk, err = lSys.ScanDisk(loopDev)
	if err != nil {
		t.Errorf("Failed to scan loopDev %s", loopDev)
	}

	ast.Equal(2, len(disk.Partitions), "number of found partitions incorrect")
	found1 := disk.Partitions[part1.Number]
	found3 := disk.Partitions[part3.Number]

	ast.Equal(part1, found1, "partition 1 differed")
	ast.Equal(part3, found3, "partition 3 differed")
	ast.Equal(uint64(100*MiB), disk.FreeSpaces()[0].Size(), "freespace gap wrong size")
}
