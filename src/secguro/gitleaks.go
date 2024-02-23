package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"
)

type GitleaksFinding struct {
	RuleID      string
	File        string
	StartLine   int
	EndLine     int
	StartColumn int
	EndColumn   int
	Match       string
	Commit      string
	Date        string
	Author      string
	Email       string
	Message     string
}

func convertGitleaksFindingToUnifiedFinding(gitleaksFinding GitleaksFinding) UnifiedFinding {
	commitSummary, _, _ := strings.Cut(gitleaksFinding.Message, "\n")

	return UnifiedFinding{
		Detector:           "gitleaks",
		Rule:               gitleaksFinding.RuleID,
		File:               gitleaksFinding.File,
		LineStart:          gitleaksFinding.StartLine,
		LineEnd:            gitleaksFinding.EndLine,
		ColumnStart:        gitleaksFinding.StartColumn,
		ColumnEnd:          gitleaksFinding.EndColumn,
		Match:              gitleaksFinding.Match,
		Hint:               "",
		CommitHash:         gitleaksFinding.Commit,
		CommitDate:         gitleaksFinding.Date,
		AuthorName:         gitleaksFinding.Author,
		AuthorEmailAddress: gitleaksFinding.Email,
		CommitSummary:      commitSummary,
	}
}

func getGitleaksOutputJson(gitMode bool) ([]byte, error) {
	gitleaksOutputJsonPath := dependenciesDir + "/gitleaksOutput.json"

	cmd := (func() *exec.Cmd {
		if gitMode {
			// secguro-ignore-next-line
			return exec.Command(dependenciesDir+"/gitleaks/gitleaks",
				"detect", "--report-format", "json", "--report-path", gitleaksOutputJsonPath)
		} else {
			// secguro-ignore-next-line
			return exec.Command(dependenciesDir+"/gitleaks/gitleaks",
				"detect", "--no-git", "--report-format", "json", "--report-path", gitleaksOutputJsonPath)
		}
	})()
	cmd.Dir = directoryToScan
	// Ignore error because this is expected to deliver an exit code not equal to 0 and write to stderr.
	out, _ := cmd.Output()
	if out == nil {
		return nil, errors.New("did not receive output from gitleaks")
	}

	gitleaksOutputJson, err := os.ReadFile(gitleaksOutputJsonPath)

	return gitleaksOutputJson, err
}

func getGitleaksFindingsAsUnified(gitMode bool) ([]UnifiedFinding, error) {
	gitleaksOutputJson, err := getGitleaksOutputJson(gitMode)
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
