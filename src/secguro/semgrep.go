package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
)

type Meta_SemgrepFinding struct {
	Results []SemgrepFinding
}

type SemgrepFinding struct {
	Checker_id string
	Start      SemgrepFinding_start
	Extra      SemgrepFinding_extra
	path       string
}

type SemgrepFinding_start struct {
	Col  int
	Line int
}

type SemgrepFinding_extra struct {
	Lines   string
	Message string
}

func convertSemgrepFindingToUnifiedFinding(semgrepFinding SemgrepFinding) UnifiedFinding {
	return UnifiedFinding{
		Detector: "semgrep",
		Rule:     semgrepFinding.Extra.Message,
		File:     semgrepFinding.path,
		Line:     semgrepFinding.Start.Line,
		Column:   semgrepFinding.Start.Col,
		Match:    semgrepFinding.Extra.Lines,
	}
}

func getSemgrepOutputJson() ([]byte, error) {
	semgrepOutputJsonPath := dependenciesDir + "/semgrepOutput.json"

	cmd := exec.Command("semgrep", "scan", "--json", "-o", semgrepOutputJsonPath)
	cmd.Dir = directoryToScan
	// Ignore error because this is expected to deliver an exit code not equal to 0 and write to stderr.
	out, _ := cmd.Output()
	if len(out) != 0 {
		return nil, errors.New("received unexpected output from semgrep")
	}

	semgrepOutputJson, err := os.ReadFile(semgrepOutputJsonPath)
	return semgrepOutputJson, err
}

func getSemgrepFindingsAsUnified() ([]UnifiedFinding, error) {
	semgrepOutputJson, err := getSemgrepOutputJson()
	if err != nil {
		return nil, err
	}

	var metaSemgrepFindings Meta_SemgrepFinding
	err = json.Unmarshal(semgrepOutputJson, &metaSemgrepFindings)
	if err != nil {
		return nil, err
	}

	semgrepFindings := metaSemgrepFindings.Results

	unifiedFindings := Map(semgrepFindings, convertSemgrepFindingToUnifiedFinding)
	return unifiedFindings, nil
}

func installSemgrep() error {
	cmd := exec.Command("python3", "-m", "pipx", "install", "semgrep")
	_, err := cmd.Output()
	if err != nil {
		return errors.New("Failed to install Semgrep. Make sure that python3 and pipx are installed.")
	}

	return nil
}
