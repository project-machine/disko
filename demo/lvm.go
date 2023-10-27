package main

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
	"machinerun.io/disko"
	"machinerun.io/disko/linux"
)

//nolint:gochecknoglobals
var lvmCommands = cli.Command{
	Name:  "lvm",
	Usage: "lvm commands",
	Subcommands: []*cli.Command{
		{
			Name:   "dump-vgs",
			Usage:  "Scan system and dump disko VGs.  Optionally give a vg name.",
			Action: lvmDumpVGs,
		},
	},
}

func lvmDumpVGs(c *cli.Context) error {
	var filter disko.VGFilter

	if c.Args().Len() == 0 {
		filter = func(v disko.VG) bool { return true }
	} else if c.Args().Len() == 1 {
		filter = func(v disko.VG) bool { return v.Name == c.Args().First() }
	} else {
		return fmt.Errorf("too many args. Really just want 1. Got %d", c.Args().Len())
	}

	vmgr := linux.VolumeManager()

	vgset, err := vmgr.ScanVGs(filter)
	if err != nil {
		return err
	}

	jbytes, err := json.MarshalIndent(&vgset, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jbytes))

	return nil
}
