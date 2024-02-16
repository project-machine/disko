package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"
	"machinerun.io/disko/smartpqi"
)

//nolint:gochecknoglobals
var smartpqiCommands = cli.Command{
	Name:  "smartpqi",
	Usage: "smartpqi / arcconf commands",
	Subcommands: []*cli.Command{
		{
			Name:   "dump",
			Usage:  "Dump information about smartpqi",
			Action: smartpqiDump,
		},
		{
			Name:   "disk-summary",
			Usage:  "Show information about virtual devices on system",
			Action: smartpqiDiskSummary,
		},
		{
			Name:   "list-controllers",
			Usage:  "Show the discovered controller IDs",
			Action: smartpqiListControllers,
		},
	},
}

func smartpqiListControllers(c *cli.Context) error {
	arc := smartpqi.ArcConf()
	ctrls, err := arc.List()
	if err != nil {
		return fmt.Errorf("failed to list controllers: %s", err)
	}

	fmt.Printf("Found %d controllers.", len(ctrls))
	for _, cID := range ctrls {
		fmt.Printf("Controller ID: %d\n", cID)
	}
	return nil
}

func smartpqiDiskSummary(c *cli.Context) error {
	var err error
	var ctrlNum = 1
	var ctrlArg = c.Args().First()

	if ctrlArg != "" {
		ctrlNum, err = strconv.Atoi(ctrlArg)
		if err != nil {
			return fmt.Errorf("could not convert to integer: %s", err)
		}
	}

	arc := smartpqi.ArcConf()
	ctrl, err := arc.Query(ctrlNum)

	if err != nil {
		return err
	}

	data := [][]string{{"Path", "Name", "DiskType", "RAID"}}

	for _, ld := range ctrl.LogicalDrives {
		stype := "HDD"

		if ld.IsSSD() {
			stype = "SSD"
		}

		name := ld.Name
		if ld.Name == "" {
			name = fmt.Sprintf("logicalid-%d", ld.ID)
		}

		data = append(data, []string{ld.DiskName, name, stype, ld.RAIDLevel})
	}

	printTextTable(data)

	return nil
}

func smartpqiDump(c *cli.Context) error {
	var err error
	var ctrlNum = 1
	var ctrlArg = c.Args().First()

	if ctrlArg != "" {
		ctrlNum, err = strconv.Atoi(ctrlArg)
		if err != nil {
			return fmt.Errorf("could not convert to integer: %s", err)
		}
	}

	arc := smartpqi.ArcConf()
	ctrl, err := arc.Query(ctrlNum)

	if err != nil {
		return err
	}

	jbytes, err := json.MarshalIndent(&ctrl, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", string(jbytes))

	return nil
}
