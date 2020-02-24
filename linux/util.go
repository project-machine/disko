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
	"strconv"
	"strings"
	"syscall"

	"github.com/anuvu/disko"
	"github.com/pkg/errors"
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
		default:
			return fmt.Errorf("error parsing line: %v", line)
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

	return fmt.Errorf(
		"command failed [%d]:\n cmd: %v\nout:%s\nerr%s",
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
		io.WriteString(stdin, input) // nolint:errcheck
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
