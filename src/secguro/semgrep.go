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
	Check_id string
	Start    SemgrepFinding_startAndEnd
	End      SemgrepFinding_startAndEnd
	Extra    SemgrepFinding_extra
	Path     string
}

type SemgrepFinding_startAndEnd struct {
	Col  int
	Line int
}

type SemgrepFinding_extra struct {
	Lines   string
	Message string
}

func convertSemgrepFindingToUnifiedFinding(semgrepFinding SemgrepFinding, gitMode bool) (UnifiedFinding, error) {
	gitInfo, err := getGitInfo(semgrepFinding.Path, semgrepFinding.Start.Line, gitMode)
	if err != nil {
		return UnifiedFinding{}, err
	}

	unifiedFinding := UnifiedFinding{
		Detector:           "semgrep",
		Rule:               semgrepFinding.Check_id,
		File:               semgrepFinding.Path,
		LineStart:          semgrepFinding.Start.Line,
		LineEnd:            semgrepFinding.End.Line,
		ColumnStart:        semgrepFinding.Start.Col,
		ColumnEnd:          semgrepFinding.End.Col,
		Match:              semgrepFinding.Extra.Lines,
		Hint:               semgrepFinding.Extra.Message,
		CommitHash:         gitInfo.CommitHash,
		CommitDate:         gitInfo.CommitDate,
		AuthorName:         gitInfo.AuthorName,
		AuthorEmailAddress: gitInfo.AuthorEmailAddress,
		CommitSummary:      gitInfo.CommitSummary,
	}

	return unifiedFinding, nil
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

func getSemgrepFindingsAsUnified(gitMode bool) ([]UnifiedFinding, error) {
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

	unifiedFindings, err := MapWithError(semgrepFindings,
		func(semgrepFinding SemgrepFinding) (UnifiedFinding, error) {
			return convertSemgrepFindingToUnifiedFinding(semgrepFinding, gitMode)
		})
	if err != nil {
		return make([]UnifiedFinding, 0), err
	}

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
