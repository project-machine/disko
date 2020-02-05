package disko_test

import (
	"testing"

	"github.com/anuvu/disko"
)

func TestLVTypeString(t *testing.T) {
	for _, d := range []struct {
		dtype    disko.LVType
		expected string
	}{
		{disko.THICK, "THICK"},
		{disko.THIN, "THIN"},
	} {
		found := d.dtype.String()
		if found != d.expected {
			t.Errorf("disko.LVType(%d).String() found %s, expected %s",
				d.dtype, found, d.expected)
		}
	}
}
