package main

import (
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
	},
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
