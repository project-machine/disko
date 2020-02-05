package linux

import (
	"fmt"
	"testing"
)

type myDetector struct {
	out      []byte
	err      []byte
	rc       int
	expected virtType
	logged   string
}

func (d *myDetector) detectVirt() ([]byte, []byte, int) {
	return d.out, d.err, d.rc
}

func (d *myDetector) logf(format string, a ...interface{}) {
	d.logged = fmt.Sprintf(format, a...)
}

func TestVirtType(t *testing.T) {
	tables := []myDetector{
		{[]byte("none\n"), []byte{}, 1, virtNone, ""},
		{[]byte("unknown\n"), []byte{}, 0, virtUnknown, ""},
		{[]byte("none\n"), []byte{}, 0, virtNone, ""},
		{[]byte("kvm\n"), []byte{}, 0, virtKvm, ""},
		{[]byte("error\n"), []byte{}, 0, virtError, ""},
		{[]byte("qemu\n"), []byte{}, 0, virtQemu, ""},
		{[]byte("zvm\n"), []byte{}, 0, virtZvm, ""},
		{[]byte("vmware\n"), []byte{}, 0, virtVmware, ""},
		{[]byte("microsoft\n"), []byte{}, 0, virtMicrosoft, ""},
		{[]byte("oracle\n"), []byte{}, 0, virtOracle, ""},
		{[]byte("xen\n"), []byte{}, 0, virtXen, ""},
		{[]byte("bochs\n"), []byte{}, 0, virtBochs, ""},
		{[]byte("uml\n"), []byte{}, 0, virtUml, ""},
		{[]byte("parallels\n"), []byte{}, 0, virtParallels, ""},
		{[]byte("bhyve\n"), []byte{}, 0, virtBhyve, ""},
		{[]byte("not-known-yet\n"), []byte{}, 0, virtUnknown, ""},
		{[]byte("unexpected\n"), []byte("error"), 3, virtError, ""},
	}

	for _, td := range tables {
		// set it back to unset so cache is not used.
		systemVirtType = virtUnset
		td := td
		found := getVirtTypeIface(&td)

		if found != td.expected {
			t.Errorf("out=%s err=%s rc=%d returned %d (%s) expected %s",
				td.out, td.err, td.rc,
				found, found.String(), td.expected.String())
		}
	}

	systemVirtType = virtOracle
	found := getVirtType()

	if found != virtOracle {
		t.Errorf("value not cached in systemVirtType. found %s expected %s\n",
			string(found), string(virtOracle))
	}
}
