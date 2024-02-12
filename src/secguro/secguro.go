package main

import (
	"errors"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

// const directoryToScan = "/home/christoph/Development/Work/wallet/"
const directoryToScan = "."

func main() {
	var flagScanGitGistory string
	var flagFormat string

	app := &cli.App{ // nolint: exhaustruct
		Flags: []cli.Flag{
			&cli.StringFlag{ // nolint: exhaustruct
				Name:        "scan-git-history",
				Value:       "true",
				Usage:       "true or false",
				Destination: &flagScanGitGistory,
			},
			&cli.StringFlag{ // nolint: exhaustruct
				Name:        "format",
				Value:       "text",
				Usage:       "text or json",
				Destination: &flagFormat,
			},
		},
		Action: func(cCtx *cli.Context) error {
			name := ""
			if cCtx.NArg() > 0 {
				name = cCtx.Args().Get(0)
			}

			if name != "scan" {
				return errors.New("unsupported command")
			}

			if cCtx.NArg() > 1 {
				return errors.New("too many commands")
			}

			if flagScanGitGistory != "true" && flagScanGitGistory != "false" {
				return errors.New("unsupported value for --scan-git-history")
			}

			if flagFormat != "text" && flagFormat != "json" {
				return errors.New("unsupported value for --format")
			}

			commandScan(flagScanGitGistory == "true", flagFormat == "json")

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
