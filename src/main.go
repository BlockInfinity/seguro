package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"secguro.com/secguro/pkg/config"
	"secguro.com/secguro/pkg/dependencycheck"
	"secguro.com/secguro/pkg/fix"
	"secguro.com/secguro/pkg/functional"
	"secguro.com/secguro/pkg/github"
	"secguro.com/secguro/pkg/login"
	"secguro.com/secguro/pkg/scan"
)

func main() { //nolint: funlen, cyclop
	// call otherwise unused function to suppress linting violations
	// due to unused symbols
	if false {
		github.CreatePrTest()
	}

	var flagGitMode bool
	var flagFormat string
	var flagOutput string
	var flagTolerance int
	var flagDisabledDetectors []string

	loginAction := func(cCtx *cli.Context) error {
		return login.CommandLogin()
	}

	flagsScanAndFixMode := []cli.Flag{
		&cli.BoolFlag{ //nolint: exhaustruct
			Name:        "git",
			Usage:       "set to scan git history and print commit information",
			Destination: &flagGitMode,
		},
		&cli.MultiStringFlag{
			Target: &cli.StringSliceFlag{ //nolint: exhaustruct
				Name:  "disabled-detectors",
				Usage: "list of detectors to disable (semgrep,gitleaks,dependencycheck)",
			},
			Value:       []string{},
			Destination: &flagDisabledDetectors,
		},
	}

	flagsOnlyScanMode := []cli.Flag{
		&cli.StringFlag{ //nolint: exhaustruct
			Name:        "format",
			Value:       "text",
			Usage:       "text or json",
			Destination: &flagFormat,
		},
		&cli.StringFlag{ //nolint: exhaustruct
			Name:        "output",
			Aliases:     []string{"o"},
			Value:       "",
			Usage:       "path to output destination",
			Destination: &flagOutput,
		},
		&cli.IntFlag{ //nolint: exhaustruct
			Name:        "tolerance",
			Value:       0,
			Usage:       "number of findings to tolerate when choosing exit code",
			Destination: &flagTolerance,
		},
	}

	scanOrFixAction := func(cCtx *cli.Context) error {
		if cCtx.NArg() > 0 {
			config.DirectoryToScan = cCtx.Args().Get(0)
		}

		if cCtx.NArg() > 1 {
			return errors.New("too many arguments")
		}

		if !functional.ArrayIncludes(flagDisabledDetectors, "dependencycheck") {
			if os.Getenv(dependencycheck.NvdApiKeyEnvVarName) == "" {
				fmt.Printf("Disabling detector dependencycheck because "+
					"environment variable %s is not set or is empty. "+
					"You may apply for an API key at: "+
					"https://nvd.nist.gov/developers/request-an-api-key\n", dependencycheck.NvdApiKeyEnvVarName)

				flagDisabledDetectors = append(flagDisabledDetectors, "dependencycheck")
			}
		}

		switch cCtx.Command.Name {
		case "scan":
			{
				if flagFormat != "text" && flagFormat != "json" {
					return errors.New("unsupported value for --format")
				}
				printAsJson := flagFormat == "json"

				err := scan.CommandScan(flagGitMode, flagDisabledDetectors,
					printAsJson, flagOutput, flagTolerance)
				if err != nil {
					return err
				}
			}
		case "fix":
			{
				err := fix.CommandFix(flagGitMode, flagDisabledDetectors)
				if err != nil {
					return err
				}
			}
		default:
			{
				return errors.New("unsupported command")
			}
		}

		return nil
	}

	app := &cli.App{ //nolint: exhaustruct
		Commands: []*cli.Command{
			{
				Name:   "login",
				Usage:  "log in to report findings to secguro web",
				Action: loginAction,
			},
			{
				Name:   "scan",
				Usage:  "scan for problems",
				Flags:  append(append([]cli.Flag{}, flagsScanAndFixMode...), flagsOnlyScanMode...),
				Action: scanOrFixAction,
			},
			{
				Name:   "fix",
				Usage:  "scan for problems and then switch to an interactive mode to fix them",
				Flags:  flagsScanAndFixMode,
				Action: scanOrFixAction,
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
