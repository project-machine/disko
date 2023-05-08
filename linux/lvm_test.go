package linux

import (
	"testing"

	"github.com/anuvu/disko"
)

//nolint:funlen
func TestLVDataToLV(t *testing.T) {
	var mySize uint64 = 10 * 1024 * 1024
	const aUUID = "iFMHAp-24c3-LENS-0IFt-4Mhj-rvhf-kBnnuS"

	for i, d := range []struct {
		input    lvmLVData
		expected disko.LV
	}{
		{
			input: lvmLVData{
				Name:   "myvol0",
				VGName: "myvg0",
				Path:   "/dev/myvg0/myvol0",
				Size:   mySize,
				UUID:   aUUID,
				Active: true,
				Pool:   "ThinDataLV",
				raw: map[string]string{
					"lv_layout": "linear",
				},
			},
			expected: disko.LV{
				Name:      "myvol0",
				Path:      "/dev/myvg0/myvol0",
				VGName:    "myvg0",
				UUID:      aUUID,
				Size:      mySize,
				Type:      disko.THICK,
				Encrypted: false,
			},
		},
		{
			input: lvmLVData{
				Name:   "myvol0",
				VGName: "myvg0",
				Path:   "/dev/myvg0/myvol0",
				Size:   mySize,
				UUID:   aUUID,
				Active: true,
				Pool:   "ThinDataLV",
				raw: map[string]string{
					"lv_layout": "thin,sparse",
				},
			},
			expected: disko.LV{
				Name:      "myvol0",
				Path:      "/dev/myvg0/myvol0",
				VGName:    "myvg0",
				UUID:      aUUID,
				Size:      mySize,
				Type:      disko.THIN,
				Encrypted: false,
			},
		},
		{
			input: lvmLVData{
				Name:   "ThinDataLV",
				VGName: "vg_ifc0",
				Path:   "",
				Size:   mySize,
				UUID:   aUUID,
				Active: true,
				Pool:   "",
				raw: map[string]string{
					"lv_layout":   "thin,pool",
					"lv_path":     "",
					"lv_dm_path":  "/dev/mapper/vg_ifc0-ThinDataLV",
					"data_lv":     "[ThinDataLV_tdata]",
					"metadata_lv": "[ThinDataLV_tmeta]",
				},
			},
			expected: disko.LV{
				Name:      "ThinDataLV",
				Path:      "",
				VGName:    "vg_ifc0",
				UUID:      aUUID,
				Size:      mySize,
				Type:      disko.THINPOOL,
				Encrypted: false,
			},
		},
	} {
		found := d.input.toLV()
		if found != d.expected {
			t.Errorf("entry %d found != expected\n%v\n%v\n", i, found, d.expected)
		}
	}
}
