package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

// const directoryToScan = "/home/christoph/Development/Work/wallet/"
const directoryToScan = "."

func main() {
	var flagScanGitGistory string
	var flagFormat string

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "scan-git-history",
				Value:       "true",
				Usage:       "true or false",
				Destination: &flagScanGitGistory,
			},
			&cli.StringFlag{
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

func commandScan(scanGitHistory bool, printAsJson bool) {
	fmt.Println("Downloading dependencies...")
	err := downloadDependencies()
	if err != nil {
		panic(err)
	}

	fmt.Println("Extracting dependencies...")
	err = extractDependencies()
	if err != nil {
		panic(err)
	}

	fmt.Println("Scanning...")
	gitleaksOutputJsonPath := dependenciesDir + "/gitleaksOutput.json"

	cmd := exec.Command(dependenciesDir+"/gitleaks/gitleaks", "detect", "--report-format", "json", "--report-path", gitleaksOutputJsonPath)
	cmd.Dir = directoryToScan
	// Ignore error because this is expected to deliver an exit code not equal to 0 and write to stderr.
	out, _ := cmd.Output()
	if out == nil {
		panic("did not receive output from gitleaks")
	}

	gitleaksOutputJson, err := os.ReadFile(gitleaksOutputJsonPath)
	if err != nil {
		panic(err)
	}

	var gitleaksFindings []GitleaksFinding
	json.Unmarshal(gitleaksOutputJson, &gitleaksFindings)

	fmt.Println("Findings:")
	unifiedFindings := Map(gitleaksFindings, convertGitleaksFindingToUnifiedFinding)

	if printAsJson {
		err = printJson(unifiedFindings)
		if err != nil {
			panic(err)
		}
	} else {
		printText(unifiedFindings)
	}

	os.Exit(0)
}

func printJson(unifiedFindings []UnifiedFinding) error {
	resultJson, err := json.Marshal(unifiedFindings)
	if err != nil {
		return err
	}

	fmt.Println(string(resultJson[:]))
	return nil
}

func printText(unifiedFindings []UnifiedFinding) {
	for i, unifiedFinding := range unifiedFindings {
		fmt.Printf("Finding %d:\n", i)
		fmt.Printf("  file: %v\n", unifiedFinding.File)
		fmt.Printf("  line: %d\n", unifiedFinding.Line)
		fmt.Printf("\n")
	}
}

type GitleaksFinding struct {
	File      string
	StartLine int
}

// The attributes need to start with capital letter because
// otherwise the JSON formatter cannot see them.
type UnifiedFinding struct {
	File string
	Line int
}

func convertGitleaksFindingToUnifiedFinding(gitleaksFinding GitleaksFinding) UnifiedFinding {
	return UnifiedFinding{
		File: gitleaksFinding.File,
		Line: gitleaksFinding.StartLine,
	}
}
