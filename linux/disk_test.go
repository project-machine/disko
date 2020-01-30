// +build linux

package linux

import (
	"testing"

	"github.com/anuvu/disko"
	"github.com/stretchr/testify/assert"
)

func TestGetAttachType(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		disko.VIRTIO,
		getAttachType(disko.UdevInfo{
			Name:       "vda",
			SysPath:    "/devices/pci0000:00/0000:00:05.0/virtio3/block/vda",
			Properties: map[string]string{},
			Symlinks:   []string{"disk/by-path/pci-0000:00:05.0"},
		}))

	assert.Equal(
		disko.ATA,
		getAttachType(disko.UdevInfo{
			Name:    "sda",
			SysPath: "/devices/pci0000:00/0000:00:0d.0/host0/target0:0:0/0:0:0:0/block/sda",
			Properties: map[string]string{
				"ID_BUS": "ata",
			},
			Symlinks: []string{"disk/by-id/ata-VBOX_HARDDISK_VB579a85b0-bf6debae"},
		}))
	assert.Equal(
		disko.USB,
		getAttachType(disko.UdevInfo{
			Name:    "sda",
			SysPath: "/devices/pci0000:00/0000:00:14.0/usb2/2-3/2-3:1.0/host0/target0:0:0/0:0:0:0/block/sda",
			Properties: map[string]string{
				"ID_BUS": "usb",
			},
			Symlinks: []string{"disk/by-id/ata-VBOX_HARDDISK_VB579a85b0-bf6debae"},
		}))
	assert.Equal(
		disko.SCSI,
		getAttachType(disko.UdevInfo{
			Name:    "sda",
			SysPath: "/devices/pci0000:00/0000:00:02.2/0000:05:00.0/host0/target0:0:8/0:0:8:0/block/sda",
			Properties: map[string]string{
				"ID_BUS": "scsi",
			},
			Symlinks: []string{"disk/by-id/scsi-35000c500a0d8963f",
				"disk/by-id/wwn-0x5000c500a0d8963f"},
		}))
	assert.Equal(
		disko.PCIE,
		getAttachType(disko.UdevInfo{
			Name:       "nvme0p1",
			SysPath:    "/devices/pci0000:00/0000:00:1c.4/0000:04:00.0/nvme/nvme0/nvme0n1",
			Properties: map[string]string{},
			Symlinks: []string{"disk/by-id/nvme-SPCC_M.2_PCIe_SSD_BD52079C067D00486555",
				"disk/by-id/nvme-eui.6479a72be0043535"},
		}))
}
