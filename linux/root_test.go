// +build linux,!skipIntegration

// nolint:errcheck,gomnd,funlen
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

type cleaner struct {
	Func    func() error
	Purpose string
}

// runLog - run command and Printf, useful for debugging errors.
func runLog(args ...string) {
	out, err, rc := runCommandWithOutputErrorRc(args...)
	fmt.Printf("%s\n", cmdString(args, out, err, rc))
}

// connect loop device, with a single partition. return the disko.Disk
func singlePartDisk(filePath string) (func() error, disko.Disk, error) {
	cleanup, loopDev, err := connectLoop(filePath)
	if err != nil {
		return cleanup, disko.Disk{}, err
	}

	lSys := linux.System()

	disk, err := lSys.ScanDisk(loopDev)
	if err != nil {
		return cleanup, disk, err
	}

	err = lSys.CreatePartition(disk,
		disko.Partition{
			Start:  disk.FreeSpaces()[0].Start,
			Last:   disk.FreeSpaces()[0].Last,
			Type:   disko.PartType(partid.LinuxLVM),
			Number: 1,
		})
	if err != nil {
		return cleanup, disk, err
	}

	disk, err = lSys.ScanDisk(disk.Path)

	return cleanup, disk, err
}

func TestRootPartition(t *testing.T) {
	iSkipOrFail(t, isRoot, canUseLoop)
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

func TestRootLVMExtend(t *testing.T) {
	iSkipOrFail(t, isRoot, canUseLoop, canUseLVM)

	ast := assert.New(t)

	var cleaners = []cleaner{}
	var pv disko.PV
	var vg disko.VG
	var lv disko.LV

	lvname := "diskotest-lv" + randStr(8)
	vgname := "diskotest-vg" + randStr(8)

	defer func() {
		for i := len(cleaners) - 1; i >= 0; i-- {
			if err := cleaners[i].Func(); err != nil {
				ast.Failf("cleanup %s: %s", cleaners[i].Purpose, err)
			}
		}
	}()

	tmpFile := getTempFile(GiB)
	cleaners = append(cleaners, cleaner{
		func() error { return os.Remove(tmpFile) },
		"remove tmpFile " + tmpFile})

	lCleanup, disk, err := singlePartDisk(tmpFile)
	cleaners = append(cleaners, cleaner{lCleanup, "singlePartdisk"})

	if err != nil {
		t.Fatalf("Failed to create a single part disk: %s", err)
	}

	lvm := linux.VolumeManager()

	pv, err = lvm.CreatePV(disk.Path + "p1")
	if err != nil {
		t.Fatalf("Failed to create pv on %s: %s\n", disk.Path, err)
	}

	cleaners = append(cleaners, cleaner{func() error { return lvm.DeletePV(pv) }, "remove pv"})

	vg, err = lvm.CreateVG(vgname, pv)

	if err != nil {
		t.Fatalf("Failed to create %s with %s: %s", vgname, pv.Path, err)
	}

	ast.Equal(vgname, vg.Name)

	cleaners = append(cleaners, cleaner{func() error { return lvm.RemoveVG(vgname) }, "remove VG"})

	var size1, size2 uint64 = disko.ExtentSize * 3, disko.ExtentSize * 5

	lv, err = lvm.CreateLV(vgname, lvname, size1, disko.THICK)
	if err != nil {
		t.Fatalf("Failed to create lv %s/%s: %s", vgname, lvname, err)
	}

	cleaners = append(cleaners,
		cleaner{func() error { return lvm.RemoveLV(vgname, lvname) }, "remove LV"})

	ast.Equal(lvname, lv.Name)
	ast.Equal(size1, lv.Size)

	vgs, errScan := lvm.ScanVGs(func(v disko.VG) bool { return v.Name == vgname })
	if errScan != nil {
		ast.Fail("failed scan1 volumes: %s", errScan)
	}

	foundLv := vgs[vgname].Volumes[lvname]
	ast.Equalf(size1, foundLv.Size, "initial volume size incorrect")

	if err := lvm.ExtendLV(vgname, lvname, size2); err != nil {
		t.Fatalf("Failed to extend LV %s/%s: %s", vgname, lvname, err)
	}

	vgs, errScan = lvm.ScanVGs(func(v disko.VG) bool { return v.Name == vgname })
	if errScan != nil {
		ast.Fail("failed scan1 volumes: %s", errScan)
	}

	foundLv = vgs[vgname].Volumes[lvname]
	ast.Equalf(size2, foundLv.Size, "extended volume size incorrect")
}
