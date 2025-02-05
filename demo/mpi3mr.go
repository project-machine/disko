package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"machinerun.io/disko/linux"
	"machinerun.io/disko/mpi3mr"
)

//nolint:gochecknoglobals
var mpi3mrCommands = cli.Command{
	Name:  "mpi3mr",
	Usage: "mpi3mr / storcli2 commands",
	Subcommands: []*cli.Command{
		{
			Name:   "dump",
			Usage:  "Dump information about mpi3mr",
			Action: mpi3mrDump,
		},
		{
			Name:   "disk-summary",
			Usage:  "Show information about virtual disk devices on system",
			Action: mpi3mrDiskSummary,
		},
		{
			Name:   "list-controllers",
			Usage:  "Show the discovered controllers' ID",
			Action: mpi3mrListControllers,
		},
	},
}

func mpi3mrListControllers(c *cli.Context) error {
	stor := mpi3mr.StorCli2()
	ctrls, err := stor.List()
	if err != nil {
		return errors.Errorf("failed to list controllers: %v", err)
	}

	fmt.Printf("Found %d controllers:\n", len(ctrls))
	for _, cID := range ctrls {
		fmt.Printf("Controller ID: %d\n", cID)
	}

	return nil
}

func mpi3mrDiskSummary(c *cli.Context) error {
	var err error
	var ctrlNum = 0
	var ctrlArg = c.Args().First()

	if ctrlArg != "" {
		ctrlNum, err = strconv.Atoi(ctrlArg)
		if err != nil {
			return errors.Errorf("invalid controller number: %v", err)
		}
	}

	stor := mpi3mr.StorCli2()
	ctrl, err := stor.Query(ctrlNum)
	if err != nil {
		return errors.Errorf("failed to query controller %d: %v", ctrlNum, err)
	}

	data := [][]string{{"Path", "Name", "Type", "State"}}

	for _, vd := range ctrl.VirtualDrives {
		stype := "HDD"

		if vd.IsSSD() {
			stype = "SSD"
		}

		name := vd.Name
		if vd.Name == "" {
			name = fmt.Sprintf("virtid-%s", vd.ID())
		}

		data = append(data, []string{vd.Path(), name, stype, vd.State})
	}

	for _, d := range ctrl.PhysicalDrives {
		if d.DG >= 0 {
			continue
		}

		path := ""
		if bname, err := linux.NameByDiskID(stor.DriverSysfsPath(), d.PID); err == nil {
			path = "/dev/" + bname
		}

		data = append(data, []string{path, fmt.Sprintf("diskid-%d", d.PID),
			d.Medium, d.State})
	}

	printTextTable(data)
	return nil
}

func mpi3mrDump(c *cli.Context) error {
	var err error
	var ctrlNum = 0
	var ctrlArg = c.Args().First()

	if ctrlArg != "" {
		ctrlNum, err = strconv.Atoi(ctrlArg)
		if err != nil {
			return errors.Errorf("invalid controller number: %v", err)
		}
	}

	stor := mpi3mr.StorCli2()
	ctrl, err := stor.Query(ctrlNum)
	if err != nil {
		return errors.Errorf("failed to query controller %d: %v", ctrlNum, err)
	}

	jbytes, err := json.MarshalIndent(&ctrl, "", "  ")
	if err != nil {
		return errors.Errorf("failed to marshal controller: %v", err)
	}

	fmt.Printf("%s\n", string(jbytes))
	return nil
}
