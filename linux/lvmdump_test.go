package linux

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var size1 uint64 = 27514634240
var size2 uint64 = 55029268480
var size3 = size2 * 2

func asBS(b uint64) string {
	return fmt.Sprintf("%dB", b)
}

func TestParseLvReport(t *testing.T) {
	ast := assert.New(t)
	rawStub := map[string]string{"ignore-key": "ignore-val"}

	found, err := parseLvReport([]byte(
		`{"report": [{"lv": [{
          "lv_active": "active",
          "lv_full_name": "atx_container/storage",
          "lv_name": "storage",
          "lv_path": "/dev/atx_container/storage",
          "lv_size": "` + asBS(size1) + `",
          "lv_uuid": "yY7AfO-dtWE-ROJR-f7G9-d70P-pjGF-lFfXgf",
          "vg_name": "atx_container",
          "pool_lv": ""
		}]}]}`))
	found[0].raw = rawStub

	ast.Equal(nil, err)
	ast.Equal(
		[]lvmLVData{
			{
				Name:   "storage",
				VGName: "atx_container",
				Path:   "/dev/atx_container/storage",
				Size:   size1,
				UUID:   "yY7AfO-dtWE-ROJR-f7G9-d70P-pjGF-lFfXgf",
				Active: true,
				Pool:   "",
				raw:    rawStub,
			}}, found)
}

func TestParseVgReport(t *testing.T) {
	ast := assert.New(t)
	rawStub := map[string]string{"ignore-key": "ignore-val"}
	found, err := parseVgReport([]byte(
		`{"report": [{"vg": [{
          "lv_count": "1",
          "pv_count": "1",
          "vg_free": "0B",
          "vg_name": "atx_container",
          "vg_size": "` + asBS(size2) + `",
          "vg_uuid": "pB0WKT-WukN-IAjl-Q1Lr-bLmH-Xh5x-In0V5e"
	    }]}]}`))
	found[0].raw = rawStub

	ast.Equal(nil, err)
	ast.Equal(
		[]lvmVGData{
			{
				Name: "atx_container",
				Size: size2,
				UUID: "pB0WKT-WukN-IAjl-Q1Lr-bLmH-Xh5x-In0V5e",
				Free: 0,
				raw:  rawStub,
			}}, found)
}

func TestParsePvReport(t *testing.T) {
	ast := assert.New(t)
	rawStub := map[string]string{"ignore-key": "ignore-val"}
	found, err := parsePvReport([]byte(
		`{"report": [{"pv": [{
		  "dev_size": "` + asBS(size2) + `",
		  "pv_free": "` + asBS(size3) + `",
		  "pv_mda_size": "` + asBS(size1) + `",
		  "pv_name": "/dev/vda3",
		  "pv_size": "` + asBS(size2) + `",
		  "pv_uuid": "Gf0GD0-hH0M-7x8i-9LQt-AAZm-ke5b-VfWlGR",
		  "vg_name": "vg0"
		}]}]}`))
	found[0].raw = rawStub

	ast.Equal(nil, err)
	ast.Equal(
		[]lvmPVData{
			{
				Path:         "/dev/vda3",
				VGName:       "vg0",
				Size:         size2,
				UUID:         "Gf0GD0-hH0M-7x8i-9LQt-AAZm-ke5b-VfWlGR",
				Free:         size3,
				MetadataSize: size1,
				raw:          rawStub,
			}}, found)
}
