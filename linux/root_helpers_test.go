// +build linux,integration

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
	"syscall"
	"time"
)

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
	napLen := time.Millisecond * 10 //nolint: gomnd
	startTime := time.Now()
	endTime := startTime.Add(30 * time.Second) // nolint: gomnd

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

func getTempFile(size int64) string {
	if fp, err := ioutil.TempFile("", "disko_test"); err != nil {
		panic(err)
	} else {
		name := fp.Name()
		fp.Close()
		if err := os.Truncate(name, size); err != nil {
			panic(err)
		}
		return name
	}
}

func randStr(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
