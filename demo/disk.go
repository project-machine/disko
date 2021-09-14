package main

import (
	"encoding/json"
	"fmt"

	"github.com/anuvu/disko"
	"github.com/anuvu/disko/linux"
	"github.com/anuvu/disko/partid"
	"github.com/urfave/cli/v2"
)

//nolint:gochecknoglobals
var diskCommands = cli.Command{
	Name:  "disk",
	Usage: "disk / partition commands",
	Subcommands: []*cli.Command{
		{
			Name:   "new-part",
			Usage:  "Create a new gpt partition and table",
			Action: diskNewPartition,
		},
		{
			Name:   "dump",
			Usage:  "Scan disks on the system and dump data (json)",
			Action: diskScan,
		},
		{
			Name:   "show",
			Usage:  "Scan disks on the system and dump data (human)",
			Action: diskShow,
		},
		{
			Name: "wipe",
			Usage: ("Quickly wipe disks on the system. Zero any existing " +
				"beginning and end of disk and any existing partitions"),
			Action: diskWipe,
		},
	},
}

func diskScan(c *cli.Context) error {
	var err error
	var jbytes []byte

	mysys := linux.System()
	matchAll := func(d disko.Disk) bool {
		return true
	}

	if c.Args().Len() == 1 {
		// a single argument will only output 1 disk, not an array of one disk.
		disk, err := mysys.ScanDisk(c.Args().First())
		if err != nil {
			return err
		}

		if jbytes, err = json.MarshalIndent(&disk, "", "  "); err != nil {
			return err
		}

		fmt.Printf("%s\n", string(jbytes))

		return nil
	}

	var disks disko.DiskSet
	if c.Args().Len() == 0 {
		disks, err = mysys.ScanAllDisks(matchAll)
	} else {
		disks, err = mysys.ScanDisks(matchAll, c.Args().Slice()...)
	}

	if err != nil {
		return err
	}

	if jbytes, err = json.MarshalIndent(disks, "", "  "); err != nil {
		return err
	}

	fmt.Printf("%s\n", string(jbytes))

	return nil
}

func diskShow(c *cli.Context) error {
	mysys := linux.System()
	disks, err := getDiskSet(mysys, c.Args().Slice()...)

	if err != nil {
		return err
	}

	for _, d := range disks {
		fmt.Printf("%s\n%s\n", d.String(), d.Details())
	}

	return nil
}

func getDiskSet(mysys disko.System, paths ...string) (disko.DiskSet, error) {
	matchAllSkipReadOnly := func(d disko.Disk) bool {
		if _, ok := d.Properties[disko.ReadOnly]; ok == true {
			return false
		}
		return true
	}

	if len(paths) == 0 || (len(paths) == 1 && paths[0] == "all") {
		return mysys.ScanAllDisks(matchAllSkipReadOnly)
	}

	return mysys.ScanDisks(matchAllSkipReadOnly, paths...)
}

func diskWipe(c *cli.Context) error {
	mysys := linux.System()

	disks, err := getDiskSet(mysys, c.Args().Slice()...)

	if err != nil {
		return err
	}

	for _, d := range disks {
		if err = mysys.Wipe(d); err != nil {
			return err
		}
	}

	return nil
}

func diskNewPartition(c *cli.Context) error {
	mysys := linux.System()
	fname := c.Args().First()

	if fname == "" {
		return fmt.Errorf("must provide disk/file to partition")
	}

	disk, err := mysys.ScanDisk(fname)

	if err != nil {
		return fmt.Errorf("failed to scan %s: %s", fname, err)
	}

	fs := disk.FreeSpaces()
	if len(fs) != 1 {
		return fmt.Errorf("expected 1 free space, found %d", fs)
	}

	myGUID := disko.GenGUID()

	part := disko.Partition{
		Start:  fs[0].Start,
		Last:   fs[0].Last,
		Type:   partid.LinuxLVM,
		Name:   "smoser1",
		ID:     myGUID,
		Number: uint(1),
	}

	if err := mysys.CreatePartition(disk, part); err != nil {
		return err
	}

	disk, err = mysys.ScanDisk(fname)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", disk.Details())

	return nil
}
