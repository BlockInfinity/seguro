package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
)

type GitleaksFinding struct {
	RuleID      string
	File        string
	StartLine   int
	StartColumn int
	Match       string
}

func convertGitleaksFindingToUnifiedFinding(gitleaksFinding GitleaksFinding) UnifiedFinding {
	return UnifiedFinding{
		Detector: "gitleaks",
		Rule:     gitleaksFinding.RuleID,
		File:     gitleaksFinding.File,
		Line:     gitleaksFinding.StartLine,
		Column:   gitleaksFinding.StartColumn,
		Match:    gitleaksFinding.Match,
	}
}

func getGitleaksOutputJson() ([]byte, error) {
	gitleaksOutputJsonPath := dependenciesDir + "/gitleaksOutput.json"

	cmd := exec.Command(dependenciesDir+"/gitleaks/gitleaks", "detect", "--report-format", "json", "--report-path", gitleaksOutputJsonPath)
	cmd.Dir = directoryToScan
	// Ignore error because this is expected to deliver an exit code not equal to 0 and write to stderr.
	out, _ := cmd.Output()
	if out == nil {
		return nil, errors.New("did not receive output from gitleaks")
	}

	gitleaksOutputJson, err := os.ReadFile(gitleaksOutputJsonPath)
	return gitleaksOutputJson, err
}

func getGitleaksFindingsAsUnified() ([]UnifiedFinding, error) {
	gitleaksOutputJson, err := getGitleaksOutputJson()
	if err != nil {
		return nil, err
	}

	var gitleaksFindings []GitleaksFinding
	err = json.Unmarshal(gitleaksOutputJson, &gitleaksFindings)
	if err != nil {
		return nil, err
	}

	unifiedFindings := Map(gitleaksFindings, convertGitleaksFindingToUnifiedFinding)
	return unifiedFindings, nil
}

func downloadAndExtractGitleaks() error {
	err := downloadDependency("gitleaks",
		"https://github.com/gitleaks/gitleaks/releases/download/v8.18.2/gitleaks_8.18.2_linux_x64.tar.gz")
	if err != nil {
		return err
	}

	err = extractGzDependency("gitleaks")
	return err
}
