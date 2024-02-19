package main

import (
	"errors"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

// const directoryToScan = "/home/christoph/Development/Work/wallet"
const directoryToScan = "."

func main() {
	var flagGitMode bool
	var flagFormat string
	var flagOutput string
	var flagTolerance int

	app := &cli.App{ // nolint: exhaustruct
		Commands: []*cli.Command{
			{
				Name:  "scan",
				Usage: "scan for problems",
				Flags: []cli.Flag{
					&cli.BoolFlag{ // nolint: exhaustruct
						Name:        "git",
						Usage:       "set to scan git history and print commit information",
						Destination: &flagGitMode,
					},
					&cli.StringFlag{ // nolint: exhaustruct
						Name:        "format",
						Value:       "text",
						Usage:       "text or json",
						Destination: &flagFormat,
					},
					&cli.StringFlag{ // nolint: exhaustruct
						Name:        "output",
						Aliases:     []string{"o"},
						Value:       "",
						Usage:       "path to output destination",
						Destination: &flagOutput,
					},
					&cli.IntFlag{ // nolint: exhaustruct
						Name:        "tolerance",
						Value:       0,
						Usage:       "number of findings to tolerate when choosing exit code",
						Destination: &flagTolerance,
					},
				},
				Action: func(cCtx *cli.Context) error {
					if cCtx.NArg() > 0 {
						return errors.New("too many arguments")
					}

					if flagFormat != "text" && flagFormat != "json" {
						return errors.New("unsupported value for --format")
					}

					commandScan(flagGitMode, flagFormat == "json", flagOutput, flagTolerance)
					return nil
				},
			},
		},
		Action: func(cCtx *cli.Context) error {
			return errors.New("no command or invalid command provided")
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
