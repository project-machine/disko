// nolint:errcheck
package linux_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

type cleanList struct {
	cleaners []cleaner
}

func (c *cleanList) Cleanup(t *testing.T) {
	for i := len(c.cleaners) - 1; i >= 0; i-- {
		if err := c.cleaners[i].Func(); err != nil {
			t.Errorf("cleanup %s: %s", c.cleaners[i].Purpose, err)
		}
	}
}

func (c *cleanList) Add(cl cleaner) {
	c.cleaners = append(c.cleaners, cl)
}

func (c *cleanList) AddF(f func() error, msg string) {
	c.cleaners = append(c.cleaners, cleaner{f, msg})
}

type cleaner struct {
	Func    func() error
	Purpose string
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

func runCommand(args ...string) error {
	out, err, rc := runCommandWithOutputErrorRc(args...)
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

// connectLoop - connect fname to a loop device.
//   return cleanup, devicePath, error
func connectLoop(fname string) (func() error, string, error) {
	var cmd = []string{"losetup", "--find", "--show", "--partscan", fname}
	var stdout, stderr []byte
	var rc int

	if stdout, stderr, rc = runCommandWithOutputErrorRc(cmd...); rc != 0 {
		return func() error { return nil }, "", cmdError(cmd, stdout, stderr, rc)
	}

	// chomp the trailing '\n'
	devPath := string(stdout[0 : len(stdout)-1])

	cleanup := func() error {
		return runCommand("losetup", "--detach="+devPath)
	}

	return cleanup, devPath, waitForFileSize(devPath)
}

func waitForFileSize(devPath string) error {
	fp, err := os.OpenFile(devPath, os.O_RDWR, 0)
	if err != nil {
		return err
	}

	defer fp.Close()

	diskLen := int64(0)
	napLen := time.Millisecond * 10
	startTime := time.Now()
	endTime := startTime.Add(30 * time.Second)

	for {
		if diskLen, err = fp.Seek(0, io.SeekEnd); err != nil {
			return err
		} else if diskLen != 0 {
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

func getTempDir() (cleaner, string) {
	p, err := ioutil.TempDir("", "disko_test")
	if err != nil {
		panic(err)
	}

	return cleaner{func() error { return os.RemoveAll(p) }, "remove tmpDir " + p}, p
}

func getTempFile(size int64) (cleaner, string) {
	fp, err := ioutil.TempFile("", "disko_test")
	if err != nil {
		panic(err)
	}

	name := fp.Name()
	fp.Close()

	if err := os.Truncate(name, size); err != nil {
		panic(err)
	}

	return cleaner{func() error { return os.Remove(name) }, "remove tempFile " + name}, name
}

// we don't need crypto/math random numbers to construct a random string
//nolint:gosec
func randStr(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func isRoot() error {
	uid := os.Geteuid()
	if uid == 0 {
		return nil
	}

	return fmt.Errorf("not root (euid=%d)", uid)
}

func writableCharDev(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s: did not exist", path)
		}

		return fmt.Errorf("%s: %s", path, err)
	}

	if fi.Mode()&os.ModeCharDevice != os.ModeCharDevice {
		return fmt.Errorf("%s: not a character device", path)
	}

	if err := unix.Access(path, unix.W_OK); err != nil {
		return fmt.Errorf("%s: not writable", path)
	}

	return nil
}

func hasCommand(name string) error {
	p := which(name)
	if p == "" {
		return fmt.Errorf("%s: command not present", p)
	}

	return nil
}

func which(name string) string {
	return whichSearch(name, strings.Split(os.Getenv("PATH"), ":"))
}

func whichSearch(name string, paths []string) string {
	var search []string

	if strings.ContainsRune(name, os.PathSeparator) {
		if path.IsAbs(name) {
			search = []string{name}
		} else {
			search = []string{"./" + name}
		}
	} else {
		search = []string{}
		for _, p := range paths {
			search = append(search, path.Join(p, name))
		}
	}

	for _, fPath := range search {
		if err := unix.Access(fPath, unix.X_OK); err == nil {
			return fPath
		}
	}

	return ""
}

func canUseLoop() error {
	if err := writableCharDev("/dev/loop-control"); err != nil {
		return err
	}

	return hasCommand("losetup")
}

func canUseLVM() error {
	if err := writableCharDev("/dev/mapper/control"); err != nil {
		return err
	}

	return hasCommand("lvm")
}

// iSkipOrFail - run checks
func iSkipOrFail(t *testing.T, checks ...func() error) {
	const envName = "DISKO_INTEGRATION"
	const allowSkip = "allow-skip"

	mode := os.Getenv(envName)

	switch mode {
	case "":
		mode = allowSkip
	case "run", allowSkip:
	case "skip":
		t.Skip(envName + "=" + mode)
		return // be explicit (not actually necessary)
	default:
		panic("Invalid value for " + envName + ": " + mode)
	}

	errors := []error{}

	for _, c := range checks {
		if err := c(); err != nil {
			if mode == allowSkip {
				t.Skip(err)
				return
			}

			errors = append(errors, err)
		}
	}

	if len(errors) == 0 {
		return
	}

	// mode is "run" and there are errors.
	for _, err := range errors {
		t.Error(err)
	}

	t.FailNow()
}

// nolint: gochecknoinits
func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
