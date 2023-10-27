package disko_test

import (
	"regexp"
	"testing"

	"machinerun.io/disko"
	"machinerun.io/disko/partid"
)

func TestStringRoundtrip(t *testing.T) {
	guidfmt := "^[0-9A-F]{8}-([0-9A-F]{4}-){3}[0-9A-F]{12}$"
	matcher := regexp.MustCompile(guidfmt)
	myGUID := disko.GenGUID()

	asStr := disko.GUIDToString(myGUID)

	if !matcher.MatchString(asStr) {
		t.Errorf(
			"guid %#v as a string (%s) did not match format %s",
			myGUID, asStr, guidfmt)
	}

	back, err := disko.StringToGUID(asStr)
	if err != nil {
		t.Errorf("StringToGUID failed %#v -> %s: %s)", myGUID, asStr, back)
	}

	if back != myGUID {
		t.Errorf("Round trip failed. %#v -> %#v", myGUID, back)
	}
}

func TestStringKnown(t *testing.T) {
	for _, td := range []struct {
		guid  disko.GUID
		asStr string
	}{
		{partid.LinuxFS, "0FC63DAF-8483-4772-8E79-3D69D8477DE4"},
		{disko.GUID{0x67, 0x45, 0x23, 0x1, 0xab, 0x89, 0xef, 0xcd, 0x1,
			0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
			"01234567-89AB-CDEF-0123-456789ABCDEF"},
	} {
		found := td.guid.String()

		if found != td.asStr {
			t.Errorf("GUIDToString(%#v) got %s. expected %s",
				td.guid, found, td.asStr)
		}

		back, err := disko.StringToGUID(found)
		if err != nil {
			t.Errorf("Failed StringToGUID(%#v): %s", found, err)
		}

		if td.guid != back {
			t.Errorf("StringToGuid(%s) returned %#v. expected %#v",
				found, back, td.guid)
		}
	}
}
