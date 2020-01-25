package mockos_test

import (
	"strings"
	"testing"

	"github.com/anuvu/disko"
	"github.com/anuvu/disko/mockos"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPV(t *testing.T) {
	Convey("testing lvm PVs", t, func() {
		sys := mockos.System("testdata/model_sys.json")
		So(sys, ShouldNotBeNil)

		lvm := mockos.LVM(sys)
		So(lvm, ShouldNotBeNil)
		pvs, err := lvm.ScanPVs(func(f disko.PV) bool { return true })
		So(err, ShouldBeNil)
		So(pvs, ShouldBeEmpty)

		_, err = lvm.CreatePV("sdxx")
		So(err, ShouldBeError)

		pv, err := lvm.CreatePV("sda")
		So(err, ShouldBeNil)
		So(pv.Name, ShouldEqual, "sda")
		So(lvm.HasPV("sda"), ShouldBeTrue)

		_, err = lvm.CreatePV("sda")
		So(err, ShouldBeError)

		err = lvm.DeletePV(disko.PV{Name: "blah"})
		So(err, ShouldBeError)

		err = lvm.DeletePV((pv))
		So(err, ShouldBeNil)
	})
}

func TestVG(t *testing.T) {
	Convey("testing lvm VGs", t, func() {
		sys := mockos.System("testdata/model_sys.json")
		lvm := mockos.LVM(sys)

		// Create a partition per disk and a PV
		disks, _ := sys.ScanAllDisks(nil)
		for _, d := range disks {
			name := d.Name + "1"
			err := sys.CreatePartition(d, disko.Partition{
				Name:   name,
				Number: 1,
				Start:  0,
				End:    d.Size,
				Type:   "ext4",
			})

			So(err, ShouldBeNil)

			_, err = lvm.CreatePV(name)
			So(err, ShouldBeNil)
		}

		// Scan all SSDs
		ssds, err := sys.ScanAllDisks(func(d disko.Disk) bool {
			return d.Type == disko.SSD && d.Attachment == disko.RAID
		})
		So(err, ShouldBeNil)
		So(ssds, ShouldNotBeEmpty)

		pvs, err := lvm.ScanPVs(func(p disko.PV) bool {
			name := strings.TrimSuffix(p.Name, "1")
			if _, ok := ssds[name]; ok {
				return true
			}
			return false
		})
		So(err, ShouldBeNil)
		So(pvs, ShouldNotBeEmpty)

		pvlist := []disko.PV{}
		for _, pv := range pvs {
			pvlist = append(pvlist, pv)
		}

		// Should be able to create a new vg
		vg, err := lvm.CreateVG("ssd0", pvlist...)
		So(err, ShouldBeNil)
		So(vg.Name, ShouldEqual, "ssd0")

		// Cannot create an existing vg
		_, err = lvm.CreateVG("ssd0", pvlist...)
		So(err, ShouldBeError)

		// Cannot create an existing vg with same pv
		_, err = lvm.CreateVG("ssd1", pvlist...)
		So(err, ShouldBeError)
	})
}
