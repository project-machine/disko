package mockos_test

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"machinerun.io/disko"
	"machinerun.io/disko/mockos"
	"machinerun.io/disko/partid"
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

//nolint:funlen
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
				Last:   d.Size,
				Type:   partid.LinuxFS,
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

		// Scan all PVs
		allPvs, err := lvm.ScanPVs(nil)
		So(err, ShouldBeNil)
		So(allPvs, ShouldNotBeEmpty)

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

		// No vgs unless we create one
		vgs, err := lvm.ScanVGs(nil)
		So(err, ShouldBeNil)
		So(vgs, ShouldBeEmpty)

		// Should be able to create a new vg
		vg, err := lvm.CreateVG("ssd0", pvlist...)
		So(err, ShouldBeNil)
		So(vg.Name, ShouldEqual, "ssd0")
		So(lvm.HasVG("ssd0"), ShouldBeTrue)
		vgs, err = lvm.ScanVGs(func(v disko.VG) bool { return vg.Name == "ssd0" })
		So(err, ShouldBeNil)
		So(len(vgs), ShouldEqual, 1)

		// Deleting PV should fail
		for _, pv := range pvlist {
			So(lvm.DeletePV(pv), ShouldBeError)
		}

		// Cannot create an existing vg
		_, err = lvm.CreateVG("ssd0", pvlist...)
		So(err, ShouldBeError)

		// Cannot create an existing vg with same pv
		_, err = lvm.CreateVG("ssd1", pvlist...)
		So(err, ShouldBeError)

		// Cannot extend a no existing vg
		err = lvm.ExtendVG("ssdaaa", pvlist...)
		So(err, ShouldBeError)

		// Cannot extend a vg with pvs already in use
		err = lvm.ExtendVG("ssd0", pvlist...)
		So(err, ShouldBeError)

		// Extend using new set of PVs
		sdaPv := allPvs["sda1"]
		So(sdaPv.Name, ShouldEqual, "sda1")
		err = lvm.ExtendVG("ssd0", sdaPv)
		So(err, ShouldBeNil)

		// Cannot remove an non existent vg
		err = lvm.RemoveVG("ssdx")
		So(err, ShouldBeError)

		// Remove the vg
		err = lvm.RemoveVG("ssd0")
		So(err, ShouldBeNil)
		So(lvm.HasVG("ssd0"), ShouldBeFalse)
	})
}

//nolint:funlen
func TestLV(t *testing.T) {
	Convey("test lvm lvs", t, func() {
		sys := mockos.System("testdata/model_sys.json")
		lvm := mockos.LVM(sys)
		So(sys, ShouldNotBeNil)
		So(lvm, ShouldNotBeNil)

		sdaPV, err := lvm.CreatePV("sda")
		So(err, ShouldBeNil)
		So(sdaPV.Name, ShouldEqual, "sda")
		sdbPV, err := lvm.CreatePV("sdb")
		So(err, ShouldBeNil)
		So(sdbPV.Name, ShouldEqual, "sdb")

		ssdVG, err := lvm.CreateVG("ssd0", sdbPV)
		So(err, ShouldBeNil)
		So(ssdVG.Name, ShouldEqual, "ssd0")

		// Cannot create lv on a unknown vg
		_, err = lvm.CreateLV("junk", "janardan", 0, disko.THICK)
		So(err, ShouldBeError)

		// Cannot create a really large LV
		_, err = lvm.CreateLV("ssd0", "lv1", uint64(0xFFFFFFFFFFFFFFFF), disko.THICK)
		So(err, ShouldBeError)

		// Create an LV
		size := ssdVG.Size / 2
		lv, err := lvm.CreateLV("ssd0", "lv1", size, disko.THICK)
		So(err, ShouldBeNil)
		So(lv.Name, ShouldEqual, "lv1")
		So(ssdVG.Volumes, ShouldNotBeEmpty)

		// Cannot create the same lv again
		_, err = lvm.CreateLV("ssd0", "lv1", size, disko.THICK)
		So(err, ShouldBeError)

		// Create another LV with remaining space
		lv, err = lvm.CreateLV("ssd0", "lv2", size, disko.THICK)
		So(err, ShouldBeNil)
		So(lv.Name, ShouldEqual, "lv2")
		So(ssdVG.Volumes, ShouldNotBeEmpty)
		So(lvm.HasLV("ssd0", "lv1"), ShouldBeTrue)

		// Cannot create any more lvs as there is no free space
		_, err = lvm.CreateLV("ssd0", "lv3", 1024, disko.THICK)
		So(err, ShouldBeError)

		// Cannot remove an non-existent lv
		So(lvm.RemoveLV("ssd0", "moon"), ShouldBeError)
		So(lvm.RemoveLV("sun", "moon"), ShouldBeError)

		// Rename the second LV to lvRenamed
		So(lvm.RenameLV("ssd0", "lv2", "lvRenamed"), ShouldBeNil)

		// Remove the second LV
		So(lvm.RemoveLV("ssd0", "lvRenamed"), ShouldBeNil)

		// Cannot extend LV that doesnt exist
		So(lvm.ExtendLV("sun", "moon", 1024), ShouldBeError)

		// Cannot extend LV to smaller size
		So(lvm.ExtendLV("ssd0", "lv1", 1024), ShouldBeError)

		// Extend LV to full size
		So(lvm.ExtendLV("ssd0", "lv1", ssdVG.Size), ShouldBeNil)

		// Cannot extend any more
		So(lvm.ExtendLV("ssd0", "lv1", ssdVG.Size+1024), ShouldBeError)

		ssdVGf := func() disko.VG {
			vgs, _ := lvm.ScanVGs(func(v disko.VG) bool { return v.Name == "ssd0" })
			return vgs["ssd0"]
		}
		So(ssdVGf().Name, ShouldEqual, "ssd0")

		// Cannot encrypt non-existent LV
		So(lvm.CryptFormat("sun", "moon", "seemysecret"), ShouldBeError)

		// Cannot crypt open an unencrypted lv
		So(lvm.CryptOpen("ssd0", "lv1", "lv1_enc", "suttisecret"), ShouldBeError)
		So(lvm.CryptOpen("sun", "moon", "jupiter", "crapsecret"), ShouldBeError)
		So(lvm.CryptClose("ssd0", "lv1", "lv1_enc"), ShouldBeError)

		// crypt format an lv
		So(lvm.CryptFormat("ssd0", "lv1", "abcdedfgh"), ShouldBeNil)
		So(ssdVGf().Volumes["lv1"].Encrypted, ShouldBeTrue)

		// Crypt open the lv
		So(lvm.CryptOpen("ssd0", "lv1", "lv1_enc", "abcdedfgh"), ShouldBeNil)
		So(ssdVGf().Volumes["lv1"].DecryptedLVName, ShouldEqual, "lv1_enc")
		So(ssdVGf().Volumes["lv1"].DecryptedLVPath, ShouldEqual, "/dev/mapper/lv1_enc")

		// Crypt close the lv
		So(lvm.CryptClose("ssd0", "lv1", "lv1_enc"), ShouldBeNil)
		So(ssdVGf().Volumes["lv1"].DecryptedLVName, ShouldEqual, "")
		So(ssdVGf().Volumes["lv1"].DecryptedLVPath, ShouldEqual, "")

		// Cannot close an unopened lv
		So(lvm.CryptClose("blah", "blee", "blee_enc"), ShouldBeError)
		So(lvm.CryptClose("ssd0", "lv1", "lv1_enc"), ShouldBeError)
	})
}
