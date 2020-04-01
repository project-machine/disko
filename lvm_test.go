package disko_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/anuvu/disko"
)

var valid = map[string]disko.LVType{
	"THICK":    disko.THICK,
	"THIN":     disko.THIN,
	"THINPOOL": disko.THINPOOL,
}

func TestLVTypeString(t *testing.T) {
	for asStr, ltype := range valid {
		found := ltype.String()
		if found != asStr {
			t.Errorf("disko.LVType(%d).String() found %s, expected %s",
				ltype, found, asStr)
		}
	}
}

func TestLVTypeJsonSerialize(t *testing.T) {
	for asStr, ltype := range valid {
		ltype := ltype

		jbytes, err := json.Marshal(&ltype)
		if err != nil {
			t.Errorf("Failed to marshal %#v: %s", ltype, err)
			continue
		}

		jstr := string(jbytes)
		if !strings.Contains(jstr, asStr) {
			t.Errorf("Did not find string ID '%s' in json: %s", asStr, jstr)
		}
	}
}

func TestLVTypeJsonUnSerialize(t *testing.T) {
	var found disko.LVType

	for asStr, ltype := range valid {
		// "4" (no quotes) is valid json rep of int 4.  "string" is rep of string.
		validJsons := []string{fmt.Sprintf("%d", ltype), "\"" + asStr + "\""}
		for _, jsonBlob := range validJsons {
			err := json.Unmarshal([]byte(jsonBlob), &found)
			if err != nil {
				t.Errorf("Failed to unmarshal %s: %s", jsonBlob, err)
			} else if found != ltype {
				t.Errorf("Unserialized %s, got %d, expected %d", jsonBlob, found, ltype)
			}
		}
	}
}
