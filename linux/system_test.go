package linux

import (
	"testing"

	"github.com/anuvu/disko"
)

func TestGetDiskProperties(t *testing.T) {
	azureSys := ("/devices/LNXSYSTM:00/LNXSYBUS:00/PNP0A03:00/device:07/VMBUS:01" +
		"/00000000-0001-8899-0000-000000000000/host1/target1:0:1/1:0:1:0/block/sdb")
	scsiSys := "/devices/pci0000:00/0000:00:02.2/0000:05:00.0/host0/target0:0:8/0:0:8:0/block/sda"

	tables := []struct {
		info     disko.UdevInfo
		expected disko.PropertySet
	}{
		{
			disko.UdevInfo{
				Name:     "sda",
				SysPath:  scsiSys,
				Symlinks: []string{},
				Properties: map[string]string{
					"ID_MODEL":    "SPCC M.2 PCIe SSD",
					"ID_REVISION": "ECFM22.6"}},
			disko.PropertySet{disko.Ephemeral: false}},
		{
			disko.UdevInfo{
				Name:     "sdb",
				SysPath:  azureSys,
				Symlinks: []string{},
				Properties: map[string]string{
					"ID_MODEL":    "SPCC M.2 PCIe SSD",
					"ID_REVISION": "ECFM22.6"}},
			disko.PropertySet{disko.Ephemeral: true}},
		{
			disko.UdevInfo{
				Name:     "sdb",
				SysPath:  azureSys,
				Symlinks: []string{},
				Properties: map[string]string{
					"DM_MULTIPATH_DEVICE_PATH": "0",
					"ID_SERIAL_SHORT":          "AWS628703BD8E5BEB551",
					"ID_WWN":                   "nvme.1d0f-4157...4616e63652053746f72616765-00000001",
					"ID_MODEL":                 "Amazon EC2 NVMe Instance Storage",
					"ID_REVISION":              "0",
					"ID_SERIAL":                "Amazon EC2 NVMe Instance Storage_AWS628703BD8E5BEB551"}},

			disko.PropertySet{disko.Ephemeral: true}},
	}

	for _, table := range tables {
		found := getDiskProperties(table.info)
		bad := []disko.Property{}

		for k, v := range table.expected {
			if found[k] != v {
				bad = append(bad, k)
			}
		}

		for k, v := range found {
			if table.expected[k] != v {
				bad = append(bad, k)
			}
		}

		if len(bad) != 0 {
			t.Errorf("getDiskProperties(%v) returned '%v'. expected '%v'",
				table.info, found, table.expected)
		}
	}
}
