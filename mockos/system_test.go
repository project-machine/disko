package mockos_test

import (
	"testing"

	"github.com/anuvu/disko"
	"github.com/anuvu/disko/mockos"
	"github.com/anuvu/disko/partid"
	. "github.com/smartystreets/goconvey/convey"
)

//nolint: funlen, gomnd
func TestSystem(t *testing.T) {
	myID, _ := disko.StringToGUID("01234567-89AB-CDEF-0123-456789ABCDEF")

	Convey("testing System Model", t, func() {
		So(func() { mockos.System("unknown") }, ShouldPanic)

		sys := mockos.System("testdata/model_sys.json")
		So(sys, ShouldNotBeNil)

		Convey("Calling ScanAllDisks with no filter function should return all the disks", func() {
			diskSet, err := sys.ScanAllDisks(func(d disko.Disk) bool { return true })
			So(err, ShouldBeNil)
			So(diskSet, ShouldNotBeEmpty)

			// ravchama and gfahimi. ScanAllDisk does not in any case return error.
			// Probably it should return error when there is no found disks to be compatible
			// with the rest of functionality
		})

		Convey("Calling ScanDisk on dev/sda path should return the disk(s) with similar path ", func() {
			disk, err := sys.ScanDisk("/dev/sda")
			So(err, ShouldBeNil)
			So(disk, ShouldNotBeNil)
			So(disk.Name, ShouldEqual, "sda")
		})

		Convey("Calling ScanDisk on path that does not contain any disk should return error ", func() {
			_, err := sys.ScanDisk("path/with/no/disk")
			So(err, ShouldNotBeNil)
		})

		Convey("Calling ScanDisk on dev/sda path should return the disk(s) with similar path", func() {
			disk, err := sys.ScanDisk("/dev/sda")
			So(err, ShouldBeNil)
			So(disk, ShouldNotBeNil)
			So(disk.Name, ShouldEqual, "sda")
		})

		Convey("Calling ScanDisks with a specific filter on dev/sda path should return corresponding disks", func() {
			disk, err := sys.ScanDisks(func(d disko.Disk) bool { return d.Size > 10000 }, "/dev/sda")
			So(err, ShouldBeNil)
			So(disk, ShouldNotBeNil)
		})

		Convey("Calling ScanDisks with a specific filter on invalid path should return error", func() {
			disk, err := sys.ScanDisks(func(d disko.Disk) bool { return d.Size > 10000 },
				"/dev/sda", "path/with/no/disk")
			So(err, ShouldNotBeNil)
			So(disk, ShouldBeNil)
		})

		Convey("Calling CreatePartition should create a partition in disk", func() {
			disk := disko.Disk{
				Name: "sda",
			}
			partition := disko.Partition{
				Start:  0,
				Last:   10000,
				ID:     myID,
				Type:   partid.LinuxFS,
				Name:   "sda1",
				Number: 1,
			}

			// CreatePartition should probably only get the name
			err := sys.CreatePartition(disk, partition)
			So(err, ShouldBeNil)

			d, _ := sys.ScanDisk("/dev/sda")
			So(len(d.Partitions), ShouldEqual, 1)
			_, ok := d.Partitions[1]
			So(ok, ShouldBeTrue)

			Convey("Calling DeletePartition should delete the partition with the specific number from a disk", func() {
				disk := disko.Disk{
					Name: "sda",
				}

				// DeletePartition should probably only get the name
				err := sys.DeletePartition(disk, 1)
				So(err, ShouldBeNil)

				err = sys.DeletePartition(disk, 10)
				So(err, ShouldBeError)

				d, _ := sys.ScanDisk("/dev/sda")
				So(len(d.Partitions), ShouldEqual, 0)

				disk.Name = "crap"
				err = sys.DeletePartition(disk, 1)
				So(err, ShouldBeError)
			})

			Convey("Calling CreatePartition with an existing partition should return error", func() {
				err := sys.CreatePartition(disk, partition)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Calling CreatePartition on a disk not being track by system should return error", func() {
			disk := disko.Disk{
				Name: "invalid",
			}
			partition := disko.Partition{
				Start:  0,
				Last:   10000,
				ID:     myID,
				Type:   partid.LinuxFS,
				Name:   "partition1",
				Number: 1,
			}
			err := sys.CreatePartition(disk, partition)
			So(err, ShouldNotBeNil)
		})
	})
}
