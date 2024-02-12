package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// The attributes need to start with capital letter because
// otherwise the JSON formatter cannot see them.
type UnifiedFinding struct {
	File string
	Line int
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
