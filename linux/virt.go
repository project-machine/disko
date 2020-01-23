// +build linux

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

func getVirtType() virtType {
	if systemVirtType != virtUnset {
		return systemVirtType
	}

	out, stderr, rc := runCommandWithOutputErrorRc("systemd-detect-virt", "--vm")
	if rc == 0 || rc == 1 {
		strOut := string(out[:len(out)-1])

		for t, s := range virtTypesToString {
			if strOut == s {
				systemVirtType = t
				break
			}
		}

		if systemVirtType == virtUnset {
			log.Printf("Unknown virt type: %s/%s", strOut, string(stderr))
		}
	} else {
		log.Printf("Failed to read virt type [%d]: %s/%s",
			rc, string(out), string(stderr))
		systemVirtType = virtError
	}

	return systemVirtType
}

func isKvm() bool {
	return getVirtType() == virtKvm
}
