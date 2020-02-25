package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "disko-demo",
		Usage: "Play around or test disko",
		Commands: []*cli.Command{
			&diskCommands,
			&megaraidCommands,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
