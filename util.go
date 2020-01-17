package disko

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
)

// UdevInfo captures the udev information about a disk.
type UdevInfo struct {
	// Name of the disk
	Name string

	// SysPath is the system path of this device.
	SysPath string

	// Symlinks for the disk.
	Symlinks []string

	// Properties is udev information as a map of key, value pairs.
	Properties map[string]string
}

// GetUdevInfo return a UdevInfo for the device with kernel name kname.
func GetUdevInfo(kname string) (UdevInfo, error) {
	out, stderr, rc := runCommandWithOutputErrorRc(
		"udevadm", "info", "--query=all", "--export", "--name="+kname)

	info := UdevInfo{Name: kname}

	if rc != 0 {
		return info,
			fmt.Errorf("error querying kname '%s' [%d]: %s", kname, rc, stderr)
	}

	return info, parseUdevInfo(out, &info)
}

func parseUdevInfo(out []byte, info *UdevInfo) error {
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

func runCommandWithOutputErrorRc(args ...string) ([]byte, []byte, int) {
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return stdout.Bytes(), stderr.Bytes(), getCommandErrorRC(err)
}

type uRange struct {
	Start, End uint64
}

func (r *uRange) Size() uint64 {
	return r.End - r.Start
}

// findRangeGaps returns a set of uRange to represent the un-used
// uint64 between min and max that are not included in ranges.
//  findRangeGaps({{10, 40}, {50, 100}}, 0, 110}) ==
//      {{0, 9}, {41, 49}, {101, 110}}
func findRangeGaps(ranges []uRange, min, max uint64) []uRange {
	// start 'ret' off with full range of min to max, then start cutting it up.
	ret := []uRange{{min, max}}

	for _, i := range ranges {
		for r := 0; r < len(ret); r++ {
			// 5 cases:
			if i.Start > ret[r].End || i.End < ret[r].Start {
				// a. i has no overlap
			} else if i.Start <= ret[r].Start && i.End >= ret[r].End {
				// b.) i is complete superset, so remove ret[r]
				ret = append(ret[:r], ret[r+1:]...)
				r--
			} else if i.Start > ret[r].Start && i.End < ret[r].End {
				// c.) i is strict subset: split ret[r]
				ret = append(
					append(ret[:r+1], uRange{i.End + 1, ret[r].End}),
					ret[r+1:]...)
				ret[r].End = i.Start - 1
				r++ // added entry is guaranteed to be 'a', so skip it.
			} else if i.Start <= ret[r].Start {
				// d.) overlap left edge to middle
				ret[r].Start = i.End + 1
			} else if i.Start <= ret[r].End {
				// e.) middle to right edge (possibly past).
				ret[r].End = i.Start - 1
			} else {
				panic(fmt.Sprintf("Error in findRangeGaps: %v, r=%d, ret=%v",
					i, r, ret))
			}
		}
	}

	return ret
}

func getBlockDevSize(dev string) (uint64, error) {
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
				"getBlockDevSize(%s): failed to convert '%s' to int", dev, d)
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
