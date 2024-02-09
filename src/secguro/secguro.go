package main

import (
	"encoding/json"
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
				log.Fatal("unsupported command")
			}

			if flagScanGitGistory != "true" && flagScanGitGistory != "false" {
				log.Fatal("unsupported value for --scan-git-history")
			}

			if flagFormat != "text" && flagFormat != "json" {
				log.Fatal("unsupported value for --format")
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
	err := downloadDependencies()
	if err != nil {
		panic(err)
	}

	err = extractDependencies()
	if err != nil {
		panic(err)
	}

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

	unifiedFindings := Map(gitleaksFindings, convertGitleaksFindingToUnifiedFinding)

	for i, unifiedFinding := range unifiedFindings {
		fmt.Printf("Finding %d:\n", i)
		fmt.Printf("  file: %v\n", unifiedFinding.file)
		fmt.Printf("  line: %d\n", unifiedFinding.line)
		fmt.Printf("\n")
	}

	os.Exit(0)
}

type GitleaksFinding struct {
	File      string
	StartLine int
}

type UnifiedFinding struct {
	file string
	line int
}

func convertGitleaksFindingToUnifiedFinding(gitleaksFinding GitleaksFinding) UnifiedFinding {
	return UnifiedFinding{
		file: gitleaksFinding.File,
		line: gitleaksFinding.StartLine,
	}
}
