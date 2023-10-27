package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"machinerun.io/disko"
	"machinerun.io/disko/linux"
	"machinerun.io/disko/partid"
	"github.com/urfave/cli/v2"
)

var version string

const sfdiskInput = `
label: gpt
unit: sectors
first-lba: 2048

1 : start=2048, size=1048576
`
const DefaultErrorCode = 127

func getCommandErrorRCDefault(err error, rcError int) int {
	if err == nil {
		return 0
	}

	exitError, ok := err.(*exec.ExitError)
	if ok {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}

	return rcError
}

func getCommandErrorRC(err error) int {
	return getCommandErrorRCDefault(err, DefaultErrorCode)
}

func cmdError(args []string, out []byte, err []byte, rc int) error {
	if rc == 0 {
		return nil
	}

	return errors.New(cmdString(args, out, err, rc))
}

func cmdString(args []string, out []byte, err []byte, rc int) string {
	tlen := len(err)
	if tlen == 0 || err[tlen-1] != '\n' {
		err = append(err, '\n')
	}

	tlen = len(out)
	if tlen == 0 || out[tlen-1] != '\n' {
		out = append(out, '\n')
	}

	return fmt.Sprintf(
		"command returned %d:\n cmd: %v\n out: %s err: %s",
		rc, args, out, err)
}

func runCommand(args ...string) error {
	out, err, rc := runCommandWithOutputErrorRc(args...)
	return cmdError(args, out, err, rc)
}

func runCommandStdin(input string, args ...string) error {
	out, err, rc := runCommandWithOutputErrorRcStdin(input, args...)
	return cmdError(args, out, err, rc)
}

func runCommandWithOutputErrorRc(args ...string) ([]byte, []byte, int) {
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return stdout.Bytes(), stderr.Bytes(), getCommandErrorRC(err)
}

func runCommandWithOutputErrorRcStdin(input string, args ...string) ([]byte, []byte, int) {
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, input) //nolint:errcheck
	}()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

	return stdout.Bytes(), stderr.Bytes(), getCommandErrorRC(err)
}

func pathExists(d string) bool {
	_, err := os.Stat(d)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

type qemuImageInfo struct {
	Path        string `json:"filename"`
	Format      string `json:"format"`
	VirtualSize int64  `json:"virtual-size"`
	ActualSize  int64  `json:"actual-size"`
	Dirty       bool   `json:"dirty-flag"`
}

func getImageInfo(fname string) (qemuImageInfo, error) {
	var imgInfo = qemuImageInfo{}

	args := []string{"qemu-img", "info", "--output=json", fname}
	stdout, stderr, rc := runCommandWithOutputErrorRc(args...)

	if rc != 0 {
		return imgInfo, cmdError(args, stdout, stderr, rc)
	}

	err := json.Unmarshal(stdout, &imgInfo)

	return imgInfo, err
}

func startQemuNBD(devPath, imgPath, imgFormat string) (func() error, error) {
	var diskLen int64
	const rcNotFound = 127

	noCleanup := func() error { return nil }

	fp, err := os.OpenFile(devPath, os.O_RDWR, 0)
	if err != nil {
		return noCleanup, err
	}

	if diskLen, err = fp.Seek(0, io.SeekEnd); err != nil {
		return noCleanup, err
	} else if diskLen != 0 {
		return noCleanup, fmt.Errorf("%s had non-zero size %d before starting", devPath, diskLen)
	}

	fp.Close()

	args := []string{"qemu-nbd", "--fork",
		// "--trace=*,file=/tmp/" + path.Base(devPath) + ".log",
		"--format=" + imgFormat, "--connect=" + devPath, imgPath}
	stdout, stderr, rc := runCommandWithOutputErrorRc(args...)

	if rc == rcNotFound {
		return noCleanup, fmt.Errorf("cmd failed.  Do you have qemu-nbd?")
	} else if rc != 0 {
		return noCleanup,
			fmt.Errorf("failed to start qemu-nbd on %s: %s",
				devPath, cmdError(args, stdout, stderr, rc))
	}

	cleanup := func() error {
		errList := []error{}

		for _, cmd := range [][]string{
			{"udevadm", "settle", "--timeout=60"},
			{"qemu-nbd", "--disconnect", devPath},
		} {
			if err := runCommand(cmd...); err != nil {
				errList = append(errList, err)
			}
		}

		for _, err := range errList {
			msgf("ERROR: %s\n", err)
		}

		if len(errList) != 0 {
			return errList[0]
		}

		return nil
	}

	return cleanup, nil
}

func waitForFileSize(devPath string) error {
	fp, err := os.OpenFile(devPath, os.O_RDWR, 0)
	if err != nil {
		return err
	}

	defer fp.Close()

	const waitTime, napLen = 30 * time.Second, 10 * time.Millisecond

	diskLen := int64(0)
	startTime := time.Now()
	endTime := startTime.Add(waitTime)

	for {
		if diskLen, err = fp.Seek(0, io.SeekEnd); err != nil {
			return err
		} else if diskLen != 0 {
			msgf("found %s length %d after %v\n", devPath, diskLen, time.Since(startTime))
			return nil
		}

		time.Sleep(napLen)

		if time.Now().After(endTime) {
			break
		}
	}

	return fmt.Errorf("gave up waiting after %v for non-zero length in %s",
		time.Since(startTime), devPath)
}

func connectDevice(fname string, imgFormat string) (func() error, string, error) {
	var cleanup = func() error { return nil }
	var devPath string
	var err error
	var forceNBD bool

	val := os.Getenv("FORCE_NBD")
	if val != "" {
		if val == "true" {
			forceNBD = true
		} else if val == "false" {
			forceNBD = false
		} else {
			return cleanup, "", fmt.Errorf("invalid value for FORCE_NBD '%s'.  Allowed: true false", val)
		}
	}

	if !forceNBD && imgFormat == "raw" {
		cleanup, devPath, err = connectLoop(fname)
	} else {
		cleanup, devPath, err = connectNBD(fname, imgFormat)
	}

	if err != nil {
		return cleanup, devPath, err
	}

	if err = waitForFileSize(devPath); err != nil {
		if errC := cleanup(); errC != nil {
			msgf("%s\n", errC)
		}

		return func() error { return nil }, devPath, err
	}

	return cleanup, devPath, nil
}

func connectLoop(fname string) (func() error, string, error) {
	nilFunc := func() error { return nil }
	loopDev0 := "/dev/loop0"
	devPath := ""

	var err error

	if !pathExists(loopDev0) {
		if err = runCommand("modprobe", "loop"); err != nil {
			return nilFunc, "", fmt.Errorf("no %s. An attempt to 'modprobe loop' failed", loopDev0)
		}

		if !pathExists(loopDev0) {
			return nilFunc, "", fmt.Errorf("no %s. modprobed loop, still none", loopDev0)
		}
	}

	var cmd = []string{"losetup", "--find", "--show", "--partscan", fname}
	var stdout, stderr []byte
	var rc int

	if stdout, stderr, rc = runCommandWithOutputErrorRc(cmd...); rc != 0 {
		return nilFunc, "", cmdError(cmd, stdout, stderr, rc)
	}

	// chomp the trailing '\n'
	devPath = string(stdout[0 : len(stdout)-1])

	return func() error {
		return runCommand("losetup", "--detach="+devPath)
	}, devPath, nil
}

func connectNBD(fname string, imgFormat string) (func() error, string, error) {
	nilFunc := func() error { return nil }
	nbdGlob := "/sys/block/nbd*"

	nbdPaths, err := filepath.Glob(nbdGlob)
	if err != nil {
		return nilFunc, "", err
	}

	if len(nbdPaths) == 0 {
		if err = runCommand("modprobe", "nbd"); err != nil {
			return nilFunc, "", fmt.Errorf("no %s. An attempt to 'modprobe nbd' failed", nbdGlob)
		}

		nbdPaths, err = filepath.Glob(nbdGlob)
		if err != nil {
			return nilFunc, "", err
		}

		msgf("modprobed nbd, now have %d nbd devs\n", len(nbdPaths))

		if len(nbdPaths) == 0 {
			return nilFunc, "",
				fmt.Errorf("no %s. 'modprobe nbd' ran but didn't fix it", nbdGlob)
		}
	}

	// select a /sys/block/nbd* that does not have a 'pid' file.
	for _, p := range nbdPaths {
		if !pathExists(p + "/pid") {
			devPath := "/dev/" + path.Base(p)
			cleanup, err := startQemuNBD(devPath, fname, imgFormat)

			return cleanup, devPath, err
		}
	}

	return nilFunc, "", fmt.Errorf("did not find available nbd device")
}

func main() {
	const maxPtSizeMiB = 10 * 1024

	app := &cli.App{
		Name:    "ptimg",
		Usage:   "Partition and use free space in a disk image.",
		Version: version,
		Action:  partCreate,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "pt-name",
				Value: "atx-data",
				Usage: "Use this name for partition name and filesystem label",
			},
			&cli.StringFlag{
				Name:  "copy",
				Value: "",
				Usage: "Copy content from DIR to new fs",
			},
			&cli.BoolFlag{
				Name:  "execute",
				Value: false,
				Usage: "Execute remaining arguments after disk creation",
			},
			&cli.BoolFlag{
				Name:  "skip-mkfs",
				Value: false,
				Usage: "Do not do the mkfs operation.",
			},
			&cli.BoolFlag{
				Name:  "skip-mount",
				Value: false,
				Usage: "Do not mount the partition.",
			},
			&cli.IntFlag{
				Name:  "max-size",
				Value: maxPtSizeMiB,
				Usage: "Maximum size in Mebibytes for partition",
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func msgf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func createOrFindPartition(dSys disko.System, devPath string, ptname string, maxSize int) (uint, error) {
	var err error
	var ptNum uint
	const size100M = 100 * disko.Mebibyte

	disk, err := dSys.ScanDisk(devPath)
	if err != nil {
		return ptNum, err
	}

	for _, p := range disk.Partitions {
		if p.Name == ptname {
			ptNum = p.Number
			msgf("Using existing partition with name=%s number=%d\n", ptname, ptNum)

			return p.Number, nil
		}
	}

	freeList := disk.FreeSpacesWithMin(size100M)
	if len(freeList) == 0 {
		return 0, fmt.Errorf(
			"could not find freespace on %s.  You can add space to it with "+
				"'qemu-img resize %s +10G'", disk.Path, disk.Path)
	}

	for i := uint(1); i < 128; i++ {
		if _, exists := disk.Partitions[i]; !exists {
			ptNum = i
			break
		}
	}

	if ptNum == 0 {
		return 0, fmt.Errorf("failed to find an open partition number")
	}

	last := freeList[0].Last
	if (last - freeList[0].Start) > uint64(maxSize)*disko.Mebibyte {
		last = freeList[0].Start + uint64(maxSize)*disko.Mebibyte - 1
	}

	newPart := disko.Partition{
		Start:  freeList[0].Start,
		Last:   last,
		Type:   partid.LinuxFS,
		Name:   ptname,
		Number: ptNum,
	}

	msgf("Creating new partition with name=%s number=%d\n", ptname, ptNum)

	err = dSys.CreatePartition(disk, newPart)

	return ptNum, err
}

func handleMount(cpSrc string, exCmd []string, devPath string, subs map[string]string) error {
	var tempDir, mountPoint string
	var err error

	if cpSrc == "" && len(exCmd) == 0 {
		return nil
	}

	cleanup := func() error {
		if mountPoint != "" {
			if err := runCommand("umount", mountPoint); err != nil {
				msgf("failed umount: %s", err)
				return err
			}
		}

		if tempDir != "" {
			if err := os.RemoveAll(tempDir); err != nil {
				msgf("failed tmpdir cleanup: %s", err)
				return err
			}
		}

		mountPoint, tempDir = "", ""

		return nil
	}

	defer cleanup() //nolint: errcheck

	if tempDir, err = ioutil.TempDir("", "atx-data-tool."); err != nil {
		return err
	}

	if err = runCommand("mount", devPath, tempDir); err != nil {
		return err
	}

	mountPoint = tempDir

	if cpSrc != "" {
		if !strings.HasSuffix(cpSrc, "/") {
			cpSrc += "/"
		}

		msgf("Copying %s -> to fs on %s mounted at %s\n", cpSrc, devPath, mountPoint)

		if err = runCommand("rsync", "--recursive", cpSrc, mountPoint); err != nil {
			return err
		}
	}

	subs["MOUNT_POINT"] = mountPoint

	return handleCommand(exCmd, subs)
}

func handleCommand(exCmd []string, subs map[string]string) error {
	if len(exCmd) == 0 {
		return nil
	}

	modCmd := []string{}
	modEnv := os.Environ()

	for _, i := range exCmd {
		for k, v := range subs {
			i = strings.ReplaceAll(i, "@"+k+"@", v)
		}

		modCmd = append(modCmd, i)
	}

	for k, v := range subs {
		modEnv = append(modEnv, k+"="+v)
	}

	cmd := exec.Command(modCmd[0], modCmd[1:]...) //nolint:gosec
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = modEnv

	msgf("Executing %v\n", modCmd)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cmd %v exited with %d", modCmd, getCommandErrorRC(err))
	}

	return nil
}

//nolint:funlen
func partCreate(c *cli.Context) error {
	var err error
	var ptNum uint
	var imgInfo qemuImageInfo

	exCmd := []string{}
	fname := c.Args().First()
	cpSrc := c.String("copy")
	skipMkfs := c.Bool("skip-mkfs")
	skipMount := c.Bool("skip-mount")
	myname := c.String("pt-name")
	maxSize := c.Int("max-size")

	if c.Bool("execute") {
		exCmd = c.Args().Slice()[1:]
	} else if c.Args().Len() != 1 {
		return fmt.Errorf("got %d arguments.  Did you mean --execute?", c.Args().Len())
	}

	if skipMount && cpSrc != "" {
		return fmt.Errorf("cannot skip mount and copy")
	}

	msgf("image = %s. cp=%s max-size=%d cmd=%v\n", fname, cpSrc, maxSize, exCmd)

	if cpSrc != "" && !pathExists(cpSrc) {
		return fmt.Errorf("--copy argument '%s' does not exist", cpSrc)
	}

	if !pathExists(fname) {
		msgf("Creating %s as qemu raw size 1GiB\n", fname)

		if err := runCommand("qemu-img", "create", "-f", "raw", fname, "1G"); err != nil {
			return err
		}

		msgf("Partitioning with sfdisk\n")

		err := runCommandStdin(sfdiskInput, "sfdisk", "--no-tell-kernel", "--no-reread", fname)
		if err != nil {
			return err
		}
	}

	imgInfo, err = getImageInfo(fname)
	if err != nil {
		return err
	}

	msgf("That image is type %s and actual=%d virtual=%d\n",
		imgInfo.Format, imgInfo.ActualSize, imgInfo.VirtualSize)

	devCleanup, devPath, err := connectDevice(imgInfo.Path, imgInfo.Format)

	defer func() {
		if err := devCleanup(); err != nil {
			msgf("Error: %s\n", err)
		}
	}()

	if err != nil {
		return err
	}

	mysys := linux.System()

	if ptNum, err = createOrFindPartition(mysys, devPath, myname, maxSize); err != nil {
		return err
	}

	disk, err := mysys.ScanDisk(devPath)
	if err != nil {
		return err
	}

	msgf("%s\n\n%s\n", disk, disk.Details())

	partPath := fmt.Sprintf("%sp%d", devPath, ptNum)

	if !skipMkfs {
		msgf("Executing mkfs.ext2 on %s\n", partPath)

		err = runCommand("mkfs.ext2", "-F", "-Elazy_itable_init=0,lazy_journal_init=0", "-L"+myname, partPath)
		if err != nil {
			return err
		}
	}

	subs := map[string]string{
		"PART_DEV":  partPath,
		"BLOCK_DEV": devPath,
	}

	if skipMount {
		return handleCommand(exCmd, subs)
	}

	return handleMount(cpSrc, exCmd, partPath, subs)
}
