package main

import (
	"fmt"
	"os"
	"path"

	"github.com/anuvu/disko"
	"github.com/anuvu/disko/linux"
	"github.com/anuvu/disko/partid"
	"github.com/urfave/cli/v2"
)

const partitionName = "updown"
const maxParts = 128
const mySecret = "passw0rd"

//nolint:gochecknoglobals
var miscCommands = cli.Command{
	Name:  "misc",
	Usage: "miscellaneous test/debug",
	Subcommands: []*cli.Command{
		{
			Name:   "updown",
			Usage:  "Create a partition table, vg, lv, take it all down",
			Action: miscUpDown,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "skip-lvm",
					Value: false,
					Usage: "Do not do lvm operations",
				},
				&cli.BoolFlag{
					Name:  "skip-pvcreate",
					Value: false,
					Usage: "Do not do pvcreate separately from vgcreate",
				},
				&cli.BoolFlag{
					Name:  "skip-partition",
					Value: false,
					Usage: "Do not create and remove partition in the loop",
				},
				&cli.BoolFlag{
					Name:  "skip-luks",
					Value: false,
					Usage: fmt.Sprintf("Do not setup luks (password is '%s')", mySecret),
				},
				&cli.BoolFlag{
					Name:  "skip-teardown",
					Value: false,
					Usage: "Do not tear down on final run - pv, vg, lv, luks will all be still up.",
				},
				&cli.IntFlag{
					Name:  "loops",
					Value: 1,
					Usage: fmt.Sprintf("Do not create a partition - requires one named '%s'", partitionName),
				},
			},
		},
	},
}

func pathExists(fpath string) bool {
	_, err := os.Stat(fpath)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

func findPartInfo(diskPath string) (disko.Partition, error) {
	const mysize = 200 * disko.Mebibyte //nolint: gomnd
	var err error

	mysys := linux.System()

	disk, err := mysys.ScanDisk(diskPath)
	if err != nil {
		return disko.Partition{}, fmt.Errorf("failed to scan %s: %s", diskPath, err)
	}

	// try to find by name (previous failed run)
	for i := uint(1); i < maxParts; i++ {
		if disk.Partitions[i].Name == partitionName {
			fmt.Printf("Using existing partition number %d with name %s.\n", i, partitionName)
			return disk.Partitions[i], nil
		}
	}

	// no --part-number given or --part-number=0 - find a number and space.
	partNum := uint(maxParts)

	for i := uint(1); i < maxParts; i++ {
		if _, ok := disk.Partitions[i]; !ok {
			partNum = i
			break
		}
	}

	if partNum == maxParts {
		return disko.Partition{}, fmt.Errorf("unable to find empty partition number on %s", disk.Path)
	}

	fss := disk.FreeSpacesWithMin(mysize)
	if len(fss) < 1 {
		return disko.Partition{}, fmt.Errorf("did not find usable freespace on %s", disk.Path)
	}

	part := disko.Partition{
		Type:   partid.LinuxLVM,
		Name:   partitionName,
		ID:     disko.GenGUID(),
		Number: partNum,
		Start:  fss[0].Start,
		Last:   fss[0].Start + mysize - 1,
	}

	return part, nil
}

//nolint: gocognit, gocyclo, funlen
func miscUpDown(c *cli.Context) error {
	fname := c.Args().First()

	var err error
	var numRuns = c.Int("loops")
	var doPartition = !c.Bool("skip-partition")
	var doCreatePV = !c.Bool("skip-pvcreate")
	var doLvm = !c.Bool("skip-lvm")
	var doLuks = !c.Bool("skip-luks")
	var skipTeardown = c.Bool("skip-teardown")
	var part disko.Partition
	var pv disko.PV
	var vg disko.VG
	var lv disko.LV

	if fname == "" {
		return fmt.Errorf("must provide disk/file to partition")
	}

	part, err = findPartInfo(fname)
	if err != nil {
		return err
	}

	mysys := linux.System()
	myvmgr := linux.VolumeManager()

	disk, err := mysys.ScanDisk(fname)

	if err != nil {
		return fmt.Errorf("failed to scan %s: %s", fname, err)
	}

	partPath := fmt.Sprintf("%s%d", disk.Path, part.Number)
	partName := fmt.Sprintf("%s%d", path.Base(disk.Path), part.Number)

	if doPartition {
		if pathExists(partPath) {
			if err = mysys.DeletePartition(disk, part.Number); err != nil {
				return err
			}

			fmt.Printf("deleted existing partition %d on %s\n", part.Number, disk.Path)
		}
	} else if !pathExists(partPath) {
		err = mysys.CreatePartition(disk, part)
		if err != nil {
			return err
		}
		delPart := func() {
			if pathExists(partPath) {
				fmt.Printf("Deleting partition %s %d\n", disk.Path, part.Number)
				if err := mysys.DeletePartition(disk, part.Number); err != nil {
					fmt.Printf("that went bad: %s\n", err)
				}
			}
		}

		defer delPart()
	}

	if disk, err = mysys.ScanDisk(fname); err != nil {
		return err
	}

	fmt.Printf("numruns=%d partition=%t createpv=%t lvm=%t luks=%t\n%s\n",
		numRuns, doPartition, doCreatePV, doLvm, doLuks, disk.Details())

	luksSuffix := "_crypt"

	for i := 0; i < numRuns; i++ {
		fmt.Printf("[%d] starting %s %d\n", i, disk.Path, part.Number)

		if doPartition {
			err = mysys.CreatePartition(disk, part)
			if err != nil {
				return err
			}

			fmt.Printf("[%d] created partition %d\n", i, part.Number)
		}

		if !pathExists(partPath) {
			fmt.Printf("partition path %s did not exist.", partPath)

			return fmt.Errorf("should have existed")
		}

		if doLvm {
			if doCreatePV {
				pv, err = myvmgr.CreatePV(partPath)
				if err != nil {
					fmt.Printf("failed to createPV(%s): %s", partPath, err)
					return err
				}

				fmt.Printf("[%d] created PV %s: %v\n", i, partPath, pv.UUID)
			} else {
				pv = disko.PV{Name: partName, Path: partPath}
			}

			vg, err = myvmgr.CreateVG("myvg0", pv)
			if err != nil {
				fmt.Printf("Failed creating vg on %s\n", pv.Name)

				return err
			}

			fmt.Printf("[%d] created VG %s\n", i, "myvg0")

			lv, err = myvmgr.CreateLV(vg.Name, "mylv0", 100*disko.Mebibyte, disko.THICK) //nolint: gomnd
			if err != nil {
				fmt.Printf("Failed creating lv %s on %s\n", "mylv0", "myvg0")

				return err
			}

			fmt.Printf("[%d] created LV %s/%s (%d) luks=%t\n",
				i, vg.Name, lv.Name, lv.Size/disko.Mebibyte, doLuks)

			luksName := vg.Name + "_" + lv.Name + luksSuffix

			if doLuks {
				if err = myvmgr.CryptFormat(vg.Name, lv.Name, mySecret); err != nil {
					fmt.Printf("Failed to CryptFormat %s/%s", vg.Name, lv.Name)
					return err
				}

				if err = myvmgr.CryptOpen(vg.Name, lv.Name, luksName, mySecret); err != nil {
					fmt.Printf("Failed to CryptOpen %s/%s %s", vg.Name, lv.Name, luksName)
					return err
				}

				fmt.Printf("[%d] created luks device %s\n", i, luksName)
			}

			if skipTeardown && i+1 == numRuns {
				fmt.Printf("Leaving everything up on final run.\n")
				continue
			}

			if doLuks {
				if err = myvmgr.CryptClose(vg.Name, lv.Name, luksName); err != nil {
					return err
				}
			}

			err = myvmgr.RemoveVG(vg.Name)
			if err != nil {
				fmt.Printf("Failed removing vg %s: %s\n", "mylv0", err)
				return err
			}

			err = myvmgr.DeletePV(pv)
			if err != nil {
				fmt.Printf("failed to DeletePV(%s): %s\n", partPath, err)
				return err
			}

			fmt.Printf("[%d] deleted PV %s\n", i, partPath)
		}

		if doPartition {
			err = mysys.DeletePartition(disk, part.Number)
			if err != nil {
				return err
			}

			if pathExists(partPath) {
				fmt.Printf("After DeletePartition %d, %s did exist.", part.Number, partPath)
				return fmt.Errorf("should not have existed")
			}

			fmt.Printf("[%d] deleted partition %s\n", i, partPath)
		}
	}

	return nil
}
