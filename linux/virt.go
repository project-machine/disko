package linux

import "log"

type virtType int

const (
	virtUnset virtType = iota
	virtUnknown
	virtNone
	virtKvm
	virtError
	virtQemu
	virtZvm
	virtVmware
	virtMicrosoft
	virtOracle
	virtXen
	virtBochs
	virtUml
	virtParallels
	virtBhyve
)

var systemVirtType = virtUnset //nolint:gochecknoglobals

var virtTypesToString = map[virtType]string{ // nolint:gochecknoglobals
	virtUnset:     "unset",
	virtUnknown:   "unknown",
	virtNone:      "none",
	virtKvm:       "kvm",
	virtError:     "error",
	virtQemu:      "qemu",
	virtZvm:       "zvm",
	virtVmware:    "vmware",
	virtMicrosoft: "microsoft",
	virtOracle:    "oracle",
	virtXen:       "xen",
	virtBochs:     "bochs",
	virtUml:       "uml",
	virtParallels: "parallels",
	virtBhyve:     "bhyve",
}

func (t *virtType) String() string {
	return virtTypesToString[*t]
}

type detector interface {
	detectVirt() ([]byte, []byte, int)
	logf(string, ...interface{})
}

type systemdDetector struct {
}

func (d *systemdDetector) detectVirt() ([]byte, []byte, int) {
	return runCommandWithOutputErrorRc("systemd-detect-virt", "--vm")
}

func (d *systemdDetector) logf(format string, a ...interface{}) {
	log.Printf(format, a...)
}

func getVirtType() virtType {
	return getVirtTypeIface(&systemdDetector{})
}

func getVirtTypeIface(sdv detector) virtType {
	if systemVirtType != virtUnset {
		return systemVirtType
	}

	out, stderr, rc := sdv.detectVirt()

	if rc == 0 || rc == 1 {
		var strOut string = ""
		if len(out) > 1 {
			strOut = string(out[:len(out)-1])
		}

		for t, s := range virtTypesToString {
			if strOut == s {
				systemVirtType = t
				break
			}
		}

		if systemVirtType == virtUnset {
			sdv.logf("Unknown virt type: %s/%s", strOut, string(stderr))

			systemVirtType = virtUnknown
		}
	} else {
		sdv.logf("Failed to read virt type [%d]: %s/%s",
			rc, string(out), string(stderr))
		systemVirtType = virtError
	}

	return systemVirtType
}

func isKvm() bool {
	return getVirtType() == virtKvm
}
