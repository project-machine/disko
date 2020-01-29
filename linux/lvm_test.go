// +build linux

package linux

import (
	"testing"

	"github.com/anuvu/disko"
	"github.com/stretchr/testify/assert"
)

func TestLVDataToLV(t *testing.T) {
	var mySize uint64 = 10 * 1024 * 1024

	assert := assert.New(t)

	lvd := lvmLVData{
		Name:   "myvol0",
		VGName: "myvg0",
		Path:   "/dev/myvg0/myvol0",
		Size:   mySize,
		UUID:   "iFMHAp-24c3-LENS-0IFt-4Mhj-rvhf-kBnnuS",
		Active: true,
		Pool:   "ThinDataLV",
		raw: map[string]string{
			"lv_layout": "linear",
		},
	}

	assert.Equal(
		lvd.toLV(),
		disko.LV{
			Name:      "myvol0",
			Path:      "/dev/myvg0/myvol0",
			VGName:    "myvg0",
			Size:      mySize,
			Type:      disko.THICK,
			Encrypted: false,
		},
	)
}
