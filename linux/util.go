package linux

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"machinerun.io/disko"
)

// GetUdevInfo return a UdevInfo for the device with kernel name kname.
func GetUdevInfo(kname string) (disko.UdevInfo, error) {
	out, stderr, rc := runCommandWithOutputErrorRc(
		"udevadm", "info", "--query=all", "--export", "--name="+kname)

	info := disko.UdevInfo{Name: kname}

	if rc != 0 {
		return info,
			fmt.Errorf("error querying kname '%s' [%d]: %s", kname, rc, stderr)
	}

	return info, parseUdevInfo(out, &info)
}

func parseUdevInfo(out []byte, info *disko.UdevInfo) error {
	var toks [][]byte
	var payload, s string
	var err error

	if info.Properties == nil {
		info.Properties = map[string]string{}
	}

	for _, line := range bytes.Split(out, []byte("\n")) {
		if len(line) == 0 {
			continue
		}

		toks = bytes.SplitN(line, []byte(": "), 2)
		payload = string(toks[1])

		switch toks[0][0] {
		case 'P':
			info.SysPath = payload
		case 'N':
			info.Name = payload
		case 'S':
			info.Symlinks = append(info.Symlinks, strings.Split(payload, " ")...)
		case 'E':
			kv := strings.SplitN(payload, "=", 2)
			// use of Unquote is to decode \x20, \x2f and friends.
			// example: ID_MODEL_ENC=Integrated\x20Camera
			// and values often have trailing whitespace.
			s, err = strconv.Unquote("\"" + kv[1] + "\"")
			if err != nil {
				return fmt.Errorf("failed to unquote %#v: %s", kv[1], err)
			}

			info.Properties[kv[0]] = strings.TrimSpace(s)
		case 'L':
			// a 'devlink priority'. skip for now.
		default:
			return fmt.Errorf("error parsing line: (%s)", line)
		}
	}

	return nil
}

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
	return getCommandErrorRCDefault(err, 127)
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

func runCommandWithOutputErrorRc(args ...string) ([]byte, []byte, int) {
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return stdout.Bytes(), stderr.Bytes(), getCommandErrorRC(err)
}

func runCommand(args ...string) error {
	out, err, rc := runCommandWithOutputErrorRc(args...)
	return cmdError(args, out, err, rc)
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

func runCommandStdin(input string, args ...string) error {
	out, err, rc := runCommandWithOutputErrorRcStdin(input, args...)
	return cmdError(args, out, err, rc)
}

func udevSettle() error {
	return runCommand("udevadm", "settle")
}

func runCommandSettled(args ...string) error {
	err := runCommand(args...)
	if err != nil {
		return err
	}

	return udevSettle()
}

func pathExists(d string) bool {
	_, err := os.Stat(d)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

func getBlockSize(dev string) (uint64, error) {
	path := path.Join("/sys/block", path.Base(dev), "queue/logical_block_size")

	content, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return uint64(0), errors.Wrapf(err, "%s did not exist: is %s a disk?", path, dev)
	}

	if err != nil {
		return uint64(0), errors.Wrapf(err, "Failed to read size for '%s'", dev)
	}

	d := strings.TrimSpace(string(content))

	v, err := strconv.Atoi(d)
	if err != nil {
		return uint64(0),
			errors.Wrapf(err,
				"getBlockSize(%s): failed to convert '%s' to int", dev, d)
	}

	return uint64(v), nil
}

func getFileSize(file *os.File) (uint64, error) {
	var err error
	var cur, pos int64

	// read the current position so we can set it back before return
	if cur, err = file.Seek(0, io.SeekCurrent); err != nil {
		return 0, err
	}

	if pos, err = file.Seek(0, io.SeekEnd); err != nil {
		return 0, err
	}

	if _, err = file.Seek(cur, io.SeekStart); err != nil {
		return 0, err
	}

	return uint64(pos), nil
}

func lvPath(vgName, lvName string) string {
	return path.Join("/dev", vgName, lvName)
}

func vgLv(vgName, lvName string) string {
	return path.Join(vgName, lvName)
}

// Ceiling returns the smallest integer equal to or larger than val that is evenly
// divisible by unit.
func Ceiling(val, unit uint64) uint64 {
	if val%unit == 0 {
		return val
	}

	return ((val + unit) / unit) * unit
}

// Floor returns the largest integer equal to or less than val that is evenly
// divisible by unit.
func Floor(val, unit uint64) uint64 {
	if val%unit == 0 {
		return val
	}

	return (val / unit) * unit
}

// IsSysPathRAID - is this sys path (udevadm info's DEVPATH) on a scsi controller.
//
//	syspath will look something like
//	   /devices/pci0000:3a/0000:3a:02.0/0000:3c:00.0/host0/target0:2:2/0:2:2:0/block/sdc
func IsSysPathRAID(syspath string, driverSysPath string) bool {
	if !strings.HasPrefix(syspath, "/sys") {
		syspath = "/sys" + syspath
	}

	if !strings.Contains(syspath, "/host") {
		return false
	}

	fp, err := filepath.EvalSymlinks(syspath)
	if err != nil {
		fmt.Printf("seriously? %s\n", err)
		return false
	}

	for _, path := range GetSysPaths(driverSysPath) {
		if strings.HasPrefix(fp, path) {
			return true
		}
	}

	return false
}

// NameByDiskID - return the linux name (sda) for the disk with given DiskID
func NameByDiskID(driverSysPath string, id int) (string, error) {
	// given ID, we expect a single file in:
	// <driverSysPath>/0000:05:00.0/host0/target0:0:<ID>/0:0:<ID>:0/block/
	// Note: This does not work for some controllers such as a MegaRAID SAS3508
	// See https://github.com/project-machine/disko/issues/101
	idStr := fmt.Sprintf("%d", id)
	blkDir := driverSysPath + "/*/host*/target0:0:" + idStr + "/0:0:" + idStr + ":0/block/*"
	matches, err := filepath.Glob(blkDir)

	if err != nil {
		return "", err
	}

	if len(matches) != 1 {
		return "", fmt.Errorf("found %d matches to %s", len(matches), blkDir)
	}

	return path.Base(matches[0]), nil
}

func GetSysPaths(driverSysPath string) []string {
	paths := []string{}
	// a raid driver has directory entries for each of the scsi hosts on that controller.
	//   $cd /sys/bus/pci/drivers/<driver name>
	//   $ for d in *; do [ -d "$d" ] || continue; echo "$d -> $( cd "$d" && pwd -P )"; done
	//    0000:3c:00.0 -> /sys/devices/pci0000:3a/0000:3a:02.0/0000:3c:00.0
	//    module -> /sys/module/<driver module name>

	// We take a hack path and consider anything with a ":" in that dir as a host path.
	matches, err := filepath.Glob(driverSysPath + "/*:*")

	if err != nil {
		fmt.Printf("errors: %s\n", err)
		return paths
	}

	for _, p := range matches {
		fp, err := filepath.EvalSymlinks(p)

		if err == nil {
			paths = append(paths, fp)
		}
	}

	return paths
}
