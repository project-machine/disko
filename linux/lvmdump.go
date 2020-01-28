// +build linux

package linux

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func readReportUint64(s string) uint64 {
	// lvm --report-format=json --unit=B puts unit 'B' at end of all sizes.
	if strings.HasSuffix(s, "B") {
		s = s[:len(s)-1]
	}

	num, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to convert string %s to int64: %s", s, err))
	}

	return num
}

type lvmPVData struct {
	Path         string
	Size         uint64
	VGName       string
	UUID         string
	Free         uint64
	MetadataSize uint64
	raw          map[string]string
}

func (d *lvmPVData) UnmarshalJSON(b []byte) error {
	var m map[string]string
	err := json.Unmarshal(b, &m)

	if err != nil {
		return err
	}

	d.raw = m
	d.Path = m["pv_name"]
	d.VGName = m["vg_name"]
	d.UUID = m["pv_uuid"]
	d.Size = readReportUint64(m["pv_size"])
	d.MetadataSize = readReportUint64(m["pv_mda_size"])
	d.Free = readReportUint64(m["pv_free"])

	return nil
}

func parsePvReport(report []byte) ([]lvmPVData, error) {
	var d map[string]([]map[string]([]lvmPVData))
	err := json.Unmarshal(report, &d)

	if err != nil {
		return []lvmPVData{}, err
	}

	return d["report"][0]["pv"], nil
}

func getPvReport() ([]lvmPVData, error) {
	out, stderr, rc := runCommandWithOutputErrorRc(
		"lvm", "pvs", "--options=pv_all,vg_name", "--report-format=json", "--unit=B")
	if rc != 0 {
		return []lvmPVData{},
			fmt.Errorf("failed lvm pvs [%d]: %s\n%s", rc, out, stderr)
	}

	return parsePvReport(out)
}
