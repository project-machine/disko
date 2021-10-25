// +build linux,!skipIntegration

// nolint:errcheck,funlen
package linux_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/anuvu/disko"
	"github.com/anuvu/disko/linux"
	"github.com/anuvu/disko/partid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

const MiB = 1024 * 1024
const GiB = MiB * 1024

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

	var cl = cleanList{}
	defer cl.Cleanup(t)

	c, tmpFile := getTempFile(GiB)
	cl.Add(c)

	if cleanup, path, err := connectLoop(tmpFile); err != nil {
		runLog("losetup", "-a")
		t.Fatalf("failed loop: %s\n", err)
	} else {
		cl.AddF(cleanup, "detach loop "+tmpFile)
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
		ID:     disko.GUID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
		Type:   disko.PartType(partid.LinuxRootX86),
		Name:   randStr(10),
		Number: 1,
	}

	// part3 leaves 100MiB gap to verify FreeSpaces
	part3 := disko.Partition{
		Start:  part1.Last + 100*MiB + 1,
		Last:   disk.FreeSpaces()[0].Last,
		ID:     disko.GUID{0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa},
		Type:   disko.PartType(partid.LinuxFS),
		Name:   randStr(8),
		Number: 3,
	}

	if err = lSys.CreatePartitions(disk, disko.PartitionSet{1: part1, 3: part3}); err != nil {
		t.Fatalf("failed create partitions %s", err)
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

	// Now add a single partition
	part2 := disko.Partition{
		Start:  disk.FreeSpaces()[0].Start,
		Last:   disk.FreeSpaces()[0].Last,
		ID:     disko.GUID{0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9},
		Type:   disko.PartType(partid.LinuxFS),
		Name:   randStr(8),
		Number: 2,
	}

	if err = lSys.CreatePartition(disk, part2); err != nil {
		t.Fatalf("failed create partition %#v", part2)
	}

	disk, err = lSys.ScanDisk(loopDev)
	if err != nil {
		t.Errorf("Failed to scan loopDev %s", loopDev)
	}

	found2 := disk.Partitions[part2.Number]
	ast.Equal(part2, found2, "partition 2 differed")
	ast.Equal(uint64(100*MiB), found2.Size(), "partition 2 had wrong size")
}

func TestRootLVMExtend(t *testing.T) {
	iSkipOrFail(t, isRoot, canUseLoop, canUseLVM)

	ast := assert.New(t)

	var cl = cleanList{}
	defer cl.Cleanup(t)

	var pv disko.PV
	var vg disko.VG
	var lv disko.LV
	var c cleaner
	var tmpFile string
	var tmpDir string

	lvname := "diskotest-lv" + randStr(8)
	vgname := "diskotest-vg" + randStr(8)

	c, tmpDir = getTempDir()
	cl.Add(c)

	c, tmpFile = getTempFile(GiB)
	cl.Add(c)

	lCleanup, disk, err := singlePartDisk(tmpFile)
	cl.AddF(lCleanup, "singlePartdisk")

	if err != nil {
		t.Fatalf("Failed to create a single part disk: %s", err)
	}

	lvm := linux.VolumeManager()

	pv, err = lvm.CreatePV(disk.Path + "p1")
	if err != nil {
		t.Fatalf("Failed to create pv on %s: %s\n", disk.Path, err)
	}

	cl.AddF(func() error { return lvm.DeletePV(pv) }, "remove pv")

	vg, err = lvm.CreateVG(vgname, pv)

	if err != nil {
		t.Fatalf("Failed to create %s with %s: %s", vgname, pv.Path, err)
	}

	ast.Equal(vgname, vg.Name)

	cl.AddF(func() error { return lvm.RemoveVG(vgname) }, "remove VG")

	var size1, size2 uint64 = disko.ExtentSize * 3, disko.ExtentSize * 5

	lv, err = lvm.CreateLV(vgname, lvname, size1, disko.THICK)
	if err != nil {
		t.Fatalf("Failed to create lv %s/%s: %s", vgname, lvname, err)
	}

	cl.AddF(func() error { return lvm.RemoveLV(vgname, lvname) }, "remove LV")

	ast.Equal(lvname, lv.Name)
	ast.Equal(size1, lv.Size)

	vgs, errScan := lvm.ScanVGs(func(v disko.VG) bool { return v.Name == vgname })
	if errScan != nil {
		ast.Fail("failed scan1 volumes: %s", errScan)
	}

	foundLv := vgs[vgname].Volumes[lvname]
	ast.Equalf(size1, foundLv.Size, "initial volume size incorrect")

	mount1 := path.Join(tmpDir, "mp1")
	os.Mkdir(mount1, 0755)

	if err := runCommand("mkfs.ext4", "-F", "-L"+lvname, foundLv.Path); err != nil {
		t.Errorf("Failed to mkfs on %s: %s", foundLv.Path, err)
	}

	if err := unix.Mount(foundLv.Path, mount1, "ext4", 0, ""); err != nil {
		t.Errorf("Failed mount: %s", err)
	}

	cl.AddF(func() error { return unix.Unmount(mount1, 0) }, "unmount lv1")

	var stat unix.Statfs_t

	if err = unix.Statfs(mount1, &stat); err != nil {
		t.Errorf("Statfs failed on mount: %s", err)
	}

	freeBefore := stat.Blocks

	if err := lvm.ExtendLV(vgname, lvname, size2); err != nil {
		t.Fatalf("Failed to extend LV %s/%s: %s", vgname, lvname, err)
	}

	if err := runCommand("resize2fs", foundLv.Path); err != nil {
		t.Error(err)
	}

	if err = unix.Statfs(mount1, &stat); err != nil {
		t.Errorf("Statfs failed on mount after: %s", err)
	}

	ast.Greater(stat.Blocks, freeBefore, "size of fs did not grow")

	vgs, errScan = lvm.ScanVGs(func(v disko.VG) bool { return v.Name == vgname })
	if errScan != nil {
		ast.Fail("failed scan1 volumes: %s", errScan)
	}

	foundLv = vgs[vgname].Volumes[lvname]
	ast.Equalf(size2, foundLv.Size, "extended volume size incorrect")
}

func runShow(args ...string) {
	out, err, rc := runCommandWithOutputErrorRc(args...)
	fmt.Print(cmdString(args, out, err, rc))
}

func TestRootLVMCreate(t *testing.T) {
	iSkipOrFail(t, isRoot, canUseLoop, canUseLVM)

	ast := assert.New(t)

	var cl = cleanList{}
	defer cl.Cleanup(t)

	var pv disko.PV
	var vg disko.VG
	var lv disko.LV
	var c cleaner
	var tmpFile string

	lvthick := "diskot-thick" + randStr(8)
	lvthinpool := "diskot-pool" + randStr(8)
	lvthin := "diskot-thin" + randStr(8)
	vgname := "diskot-vg" + randStr(8)

	c, tmpFile = getTempFile(4 * GiB)
	cl.Add(c)

	lCleanup, disk, err := singlePartDisk(tmpFile)
	cl.AddF(lCleanup, "singlePartdisk")

	if err != nil {
		t.Fatalf("Failed to create a single part disk: %s", err)
	}

	lvm := linux.VolumeManager()

	pv, err = lvm.CreatePV(disk.Path + "p1")
	if err != nil {
		t.Fatalf("Failed to create pv on %s: %s\n", disk.Path, err)
	}

	cl.AddF(func() error { return lvm.DeletePV(pv) }, "remove pv")

	vg, err = lvm.CreateVG(vgname, pv)

	if err != nil {
		t.Fatalf("Failed to create %s with %s: %s", vgname, pv.Path, err)
	}

	cl.AddF(func() error { return lvm.RemoveVG(vgname) }, "remove VG")

	ast.Equal(vgname, vg.Name)

	thickSize := uint64(12 * MiB)

	lv, err = lvm.CreateLV(vgname, lvthick, thickSize, disko.THICK)
	if err != nil {
		t.Fatalf("Failed to create lv %s/%s: %s", vgname, lvthick, err)
	}

	cl.AddF(func() error { return lvm.RemoveLV(vgname, lvthick) }, "remove LV")

	ast.Equal(lvthick, lv.Name)
	ast.Equal(thickSize, lv.Size)

	thinPoolSize, thinSize := uint64(500*MiB), uint64(200*MiB)

	// create a THINPOOL volume
	lv, err = lvm.CreateLV(vgname, lvthinpool, thinPoolSize, disko.THINPOOL)
	if err != nil {
		t.Fatalf("Failed to create lv %s/%s: %s", vgname, lvthick, err)
	}

	cl.AddF(func() error { return lvm.RemoveLV(vgname, lvthinpool) }, "remove thin pool LV")

	ast.Equal(lvthinpool, lv.Name)
	ast.Equal(thinPoolSize, lv.Size)

	lv, err = lvm.CreateLV(vgname+"/"+lvthinpool, lvthin, thinSize, disko.THIN)
	if err != nil {
		runShow("lvm", "lvdisplay", "--unit=m", vgname)
		t.Fatalf("Failed to create THIN lv %s on %s/%s: %s", lvthin, vgname, lvthinpool, err)
	}

	ast.Equal(lvthin, lv.Name)
	ast.Equal(thinSize, lv.Size)

	vgs, errScan := lvm.ScanVGs(func(v disko.VG) bool { return v.Name == vgname })
	if errScan != nil {
		t.Fatalf("Failed to scan VGs: %s\n", err)
	}

	ast.Equal(len(vgs), 1)
}

func getTempVG(size int64, cl *cleanList) (string, error) {
	var pv disko.PV
	var vg disko.VG
	var c cleaner
	var tmpFile string

	vgname := "diskot-vg" + randStr(8)

	c, tmpFile = getTempFile(size)
	cl.Add(c)

	lCleanup, disk, err := singlePartDisk(tmpFile)
	cl.AddF(lCleanup, "singlePartdisk")

	if err != nil {
		return vgname, fmt.Errorf("failed to create a single part disk: %s", err)
	}

	lvm := linux.VolumeManager()

	pv, err = lvm.CreatePV(disk.Path + "p1")
	if err != nil {
		return vgname, fmt.Errorf("failed to create pv on %s: %s", disk.Path, err)
	}

	cl.AddF(func() error { return lvm.DeletePV(pv) }, "remove pv")

	vg, err = lvm.CreateVG(vgname, pv)

	if err != nil {
		return vgname, fmt.Errorf("failed to create %s with %s: %s", vgname, pv.Path, err)
	}

	cl.AddF(func() error { return lvm.RemoveVG(vgname) }, "remove VG")

	if vgname != vg.Name {
		return vgname, fmt.Errorf("expected vgname '%s', found '%s'", vgname, vg.Name)
	}

	return vgname, nil
}

func createLV(vg string, name string, size uint64, secret string) (cleanList, error) {
	cl := cleanList{}
	lvm := linux.VolumeManager()

	lv, err := lvm.CreateLV(vg, name, size, disko.THICK)
	if err != nil {
		return cl, err
	}

	cl.AddF(func() error { return lvm.RemoveLV(vg, name) }, "remove LV")

	// Just check the newly created LV that it has "enough" zeros
	enoughZeros := 4096
	buf := make([]byte, enoughZeros)

	devFile, err := os.Open(lv.Path)
	if err != nil {
		return cl, fmt.Errorf("failed to open device '%s': %v", lv.Path, err)
	}

	rlen, err := devFile.Read(buf)
	devFile.Close()

	if err != nil {
		return cl, fmt.Errorf("failed to read from device '%s': %v", lv.Path, err)
	} else if rlen != enoughZeros {
		return cl, fmt.Errorf("Expected to read %d from device '%s', only read %d", enoughZeros, lv.Path, rlen)
	}

	for i := 0; i < enoughZeros; i++ {
		if buf[i] != 0 {
			return cl, fmt.Errorf("device '%s' did not have enough zeros: %v", lv.Path, buf)
		}
	}

	if secret != "" {
		if err := lvm.CryptFormat(vg, name, secret); err != nil {
			return cl, err
		}

		ptName := name + "_" + randStr(8)
		if err := lvm.CryptOpen(vg, name, ptName, secret); err != nil {
			return cl, err
		}

		cl.AddF(func() error { return lvm.CryptClose(vg, name, ptName) }, "close crypt "+name)
	}

	vgs, err := lvm.ScanVGs(func(v disko.VG) bool { return v.Name == vg })
	if err != nil {
		return cl, fmt.Errorf("failed scan volumes: %s", err)
	}

	foundLv, ok := vgs[vg].Volumes[name]
	if !ok {
		return cl, fmt.Errorf("Did not find vg/lv '%s/%s' in scan", vg, name)
	}

	devPath := foundLv.Path
	if secret != "" {
		devPath = foundLv.DecryptedLVPath
	}

	if err := runCommand("mkfs.ext4", "-F", "-L"+name, devPath); err != nil {
		return cl, fmt.Errorf("Failed to mkfs on %s: %s", devPath, err)
	}

	return cl, nil
}

// TestRootLVMRecreate - create, some volumes, remove them and then recreate in the
// same order.a thick volume encrypted volume with filesystem.
// remove it and then do the same.
func TestRootLVMRecreate(t *testing.T) {
	iSkipOrFail(t, isRoot, canUseLoop, canUseLVM)

	var cl = cleanList{}
	defer cl.Cleanup(t)

	vgname, err := getTempVG(4*GiB, &cl)
	if err != nil {
		t.Fatalf("Failed to get a temp VG: %v", err)
	}

	type lvInfo struct {
		Name     string
		Size     uint64
		cleanups cleanList
	}

	for _, runName := range []string{"initial", "secret", "plain"} {
		lvdatas := []*lvInfo{
			{"lv1", 128 * MiB, cleanList{}},
			{"lv2", 128 * MiB, cleanList{}},
		}

		secret := runName
		if runName == "plain" {
			secret = ""
		}

		for _, lvd := range lvdatas {
			if cl, err := createLV(vgname, lvd.Name, lvd.Size, secret); err != nil {
				cl.Cleanup(t)
				t.Fatalf("Failed create %s - %s: %v", runName, lvd.Name, err)
			} else {
				lvd.cleanups = cl
			}
		}

		for _, lvd := range lvdatas {
			lvd.cleanups.Cleanup(t)
		}
	}
}
