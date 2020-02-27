package main

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"strconv"

	"github.com/anuvu/disko/megaraid"
	"github.com/urfave/cli/v2"
)

//nolint:gochecknoglobals
var megaraidCommands = cli.Command{
	Name:  "megaraid",
	Usage: "megaraid / storcli commands",
	Subcommands: []*cli.Command{
		{
			Name:   "dump",
			Usage:  "Dump information about megaraid",
			Action: megaraidDump,
		},
		{
			Name:   "disk-summary",
			Usage:  "Show information about virtual devices on system",
			Action: megaraidDiskSummary,
		},
	},
}

func megaraidNameByDiskID(id int) (string, error) {
	// given ID, we expect a single file in:
	// <mrSysPath>/0000:05:00.0/host0/target0:0:<ID>/0:0:<ID>:0/block/
	mrSysPath := "/sys/bus/pci/drivers/megaraid_sas"
	idStr := fmt.Sprintf("%d", id)
	blkDir := mrSysPath + "/*/host*/target0:0:" + idStr + "/0:0:" + idStr + ":0/block/*"
	matches, err := filepath.Glob(blkDir)

	if err != nil {
		return "", err
	}

	if len(matches) != 1 {
		return "", fmt.Errorf("found %d matches to %s", len(matches), blkDir)
	}

	return path.Base(matches[0]), nil
}

func megaraidDiskSummary(c *cli.Context) error {
	var err error
	var ctrlNum = 0
	var ctrlArg = c.Args().First()

	if ctrlArg != "" {
		ctrlNum, err = strconv.Atoi(ctrlArg)
		if err != nil {
			return fmt.Errorf("could not convert to integer: %s", err)
		}
	}

	mraid := megaraid.StorCli()
	ctrl, err := mraid.Query(ctrlNum)

	if err != nil {
		return err
	}

	data := [][]string{{"Path", "Name", "Type", "State"}}

	for _, vd := range ctrl.VirtDrives {
		stype := "HDD"

		if ctrl.DriveGroups[vd.DriveGroup].IsSSD() {
			stype = "SSD"
		}

		data = append(data, []string{vd.Path, vd.RaidName, stype, vd.Raw["State"]})
	}

	for _, d := range ctrl.Drives {
		if d.DriveGroup >= 0 {
			continue
		}

		path := ""
		if bname, err := megaraidNameByDiskID(d.ID); err == nil {
			path = "/dev/" + bname
		}

		data = append(data, []string{path, fmt.Sprintf("diskid-%d", d.ID),
			d.MediaType.String(), d.State})
	}

	printTextTable(data)

	return nil
}

func megaraidDump(c *cli.Context) error {
	var err error
	var ctrlNum = 0
	var ctrlArg = c.Args().First()

	if ctrlArg != "" {
		ctrlNum, err = strconv.Atoi(ctrlArg)
		if err != nil {
			return fmt.Errorf("could not convert to integer: %s", err)
		}
	}

	mraid := megaraid.StorCli()
	ctrl, err := mraid.Query(ctrlNum)

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
