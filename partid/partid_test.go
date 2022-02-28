package partid_test

import (
	"testing"

	"github.com/anuvu/disko/partid"
)

func TestPartID(t *testing.T) {
	// Not a very good test, but something.
	for id, text := range map[[16]byte]string{
		partid.LinuxFS:   "Linux-FS",
		partid.LinuxLVM:  "LVM",
		partid.LinuxRAID: "RAID",
	} {
		if partid.Text[id] != text {
			t.Errorf("Unexpected text. found %s expected %s",
				partid.Text[id], text)
		}
	}
}
