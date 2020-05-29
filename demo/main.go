package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

var version string

func printTextTable(data [][]string) {
	var lengths = make([]int, len(data[0]))

	for _, line := range data {
		for i, field := range line {
			if len(field) > lengths[i] {
				lengths[i] = len(field)
			}
		}
	}

	fmts := make([]string, len(lengths))

	for i, l := range lengths {
		fmts[i] = fmt.Sprintf("%%-%ds", l)
	}

	pfmt := strings.Join(fmts, " | ") + " |\n"

	for _, line := range data {
		s := make([]interface{}, len(line))
		for i, v := range line {
			s[i] = v
		}

		fmt.Printf(pfmt, s...)
	}
}

func main() {
	app := &cli.App{
		Name:    "disko-demo",
		Version: version,
		Usage:   "Play around or test disko",
		Commands: []*cli.Command{
			&diskCommands,
			&megaraidCommands,
			&lvmCommands,
			&miscCommands,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
