package main

import (
	"encoding/json"
	"fmt"
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
	},
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
