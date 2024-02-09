package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// const directoryToScan = "/home/christoph/Development/Work/wallet/"
const directoryToScan = "."

func main() {
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

	fmt.Println(unifiedFindings)

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
