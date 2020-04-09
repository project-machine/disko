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

func getPvReport(args ...string) ([]lvmPVData, error) {
	cmd := []string{"lvm", "pvs", "--options=pv_all,vg_name", "--report-format=json", "--unit=B"}
	cmd = append(cmd, args...)
	out, stderr, rc := runCommandWithOutputErrorRc(cmd...)

	if rc != 0 {
		return []lvmPVData{},
			fmt.Errorf("failed lvm pvs [%d]: %s\n%s", rc, out, stderr)
	}

	return parsePvReport(out)
}

type lvmVGData struct {
	Name string
	Size uint64
	UUID string
	Free uint64
	raw  map[string]string
}

func (d *lvmVGData) UnmarshalJSON(b []byte) error {
	var m map[string]string
	err := json.Unmarshal(b, &m)

	if err != nil {
		return err
	}

	d.raw = m
	d.Name = m["vg_name"]
	d.Size = readReportUint64(m["vg_size"])
	d.UUID = m["vg_uuid"]
	d.Free = readReportUint64(m["vg_free"])

	return nil
}

func parseVgReport(report []byte) ([]lvmVGData, error) {
	var d map[string]([]map[string]([]lvmVGData))
	err := json.Unmarshal(report, &d)

	if err != nil {
		return []lvmVGData{}, err
	}

	return d["report"][0]["vg"], nil
}

func getVgReport(args ...string) ([]lvmVGData, error) {
	cmd := []string{"lvm", "vgs", "--options=vg_all", "--report-format=json", "--unit=B"}
	cmd = append(cmd, args...)
	out, stderr, rc := runCommandWithOutputErrorRc(cmd...)

	if rc != 0 {
		return []lvmVGData{},
			fmt.Errorf("failed lvm vgs [%d]: %s\n%s", rc, out, stderr)
	}

	return parseVgReport(out)
}

type lvmLVData struct {
	Name   string
	VGName string
	Path   string
	Size   uint64
	UUID   string
	Active bool
	Pool   string
	raw    map[string]string
}

func (d *lvmLVData) UnmarshalJSON(b []byte) error {
	var m map[string]string

	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	d.raw = m
	d.Path = m["lv_path"]
	d.Name = m["lv_name"]
	d.VGName = m["vg_name"]
	d.Active = m["lv_active"] == "active"
	d.Pool = m["pool_lv"]
	d.UUID = m["lv_uuid"]
	d.Size = readReportUint64(m["lv_size"])

	return nil
}

func parseLvReport(report []byte) ([]lvmLVData, error) {
	var d map[string]([]map[string]([]lvmLVData))

	err := json.Unmarshal(report, &d)
	if err != nil {
		return []lvmLVData{}, err
	}

	return d["report"][0]["lv"], nil
}

func getLvReport(args ...string) ([]lvmLVData, error) {
	cmd := []string{"lvm", "lvs", "--options=lv_all,vg_name", "--report-format=json", "--unit=B"}
	cmd = append(cmd, args...)
	out, stderr, rc := runCommandWithOutputErrorRc(cmd...)

	if rc != 0 {
		return []lvmLVData{},
			fmt.Errorf("failed lvm lvs [%d]: %s\n%s", rc, out, stderr)
	}

	return parseLvReport(out)
}
