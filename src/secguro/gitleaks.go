package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"runtime"
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

func convertGitleaksFindingToUnifiedFinding(gitMode bool,
	gitleaksFinding GitleaksFinding) (UnifiedFinding, error) {
	gitInfo, err := getGitInfo(gitMode, gitleaksFinding.Commit,
		gitleaksFinding.File, gitleaksFinding.StartLine, false)
	if err != nil {
		return UnifiedFinding{}, err
	}

	currentLocationGitInfo, err := getGitInfo(gitMode, gitleaksFinding.Commit,
		gitleaksFinding.File, gitleaksFinding.StartLine, true)
	if err != nil {
		return UnifiedFinding{}, err
	}

	unifiedFinding := UnifiedFinding{
		Detector:    "gitleaks",
		Rule:        gitleaksFinding.RuleID,
		File:        gitleaksFinding.File,
		LineStart:   gitleaksFinding.StartLine,
		LineEnd:     gitleaksFinding.EndLine,
		ColumnStart: gitleaksFinding.StartColumn,
		ColumnEnd:   gitleaksFinding.EndColumn,
		Match:       gitleaksFinding.Match,
		Hint:        "",
		GitInfo:     gitInfo,
	}

	if currentLocationGitInfo != nil {
		latestCommitHash, err := getLatestCommitHash()
		if err != nil {
			return UnifiedFinding{}, err
		}

		if currentLocationGitInfo.CommitHash == latestCommitHash {
			unifiedFinding.File = currentLocationGitInfo.File
			unifiedFinding.LineStart = currentLocationGitInfo.Line
			unifiedFinding.LineEnd =
				currentLocationGitInfo.Line + gitleaksFinding.EndLine - gitleaksFinding.StartLine
		} else {
			unifiedFinding.File = ""
			unifiedFinding.LineStart = -1
			unifiedFinding.LineEnd = -1
		}
	}

	return unifiedFinding, nil
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

	unifiedFindings, err := MapWithError(gitleaksFindings,
		func(gitleaksFinding GitleaksFinding) (UnifiedFinding, error) {
			return convertGitleaksFindingToUnifiedFinding(gitMode, gitleaksFinding)
		})
	if err != nil {
		return make([]UnifiedFinding, 0), err
	}

	return unifiedFindings, nil
}

func downloadAndExtractGitleaks() error {
	var url string
	switch runtime.GOOS {
	case "linux":
		url = "https://github.com/gitleaks/gitleaks/releases/download/v8.18.2/gitleaks_8.18.2_linux_x64.tar.gz"
	case "darwin":
		url = "https://github.com/gitleaks/gitleaks/releases/download/v8.18.2/gitleaks_8.18.2_darwin_arm64.tar.gz"
	default:
		return errors.New("Unsupported platform")
	}

	err := downloadDependency("gitleaks", "tar.gz", url)
	if err != nil {
		return err
	}

	err = extractGzDependency("gitleaks")

	return err
}
