package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
)

type GitleaksFinding struct {
	File      string
	StartLine int
}

func convertGitleaksFindingToUnifiedFinding(gitleaksFinding GitleaksFinding) UnifiedFinding {
	return UnifiedFinding{
		File: gitleaksFinding.File,
		Line: gitleaksFinding.StartLine,
	}
}

func getGitleaksOutputJson() ([]byte, error) {
	gitleaksOutputJsonPath := dependenciesDir + "/gitleaksOutput.json"

	cmd := exec.Command(dependenciesDir+"/gitleaks/gitleaks", "detect", "--report-format", "json", "--report-path", gitleaksOutputJsonPath)
	cmd.Dir = directoryToScan
	// Ignore error because this is expected to deliver an exit code not equal to 0 and write to stderr.
	out, _ := cmd.Output()
	if out == nil {
		errors.New("did not receive output from gitleaks")
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
	json.Unmarshal(gitleaksOutputJson, &gitleaksFindings)
	unifiedFindings := Map(gitleaksFindings, convertGitleaksFindingToUnifiedFinding)
	return unifiedFindings, nil
}
