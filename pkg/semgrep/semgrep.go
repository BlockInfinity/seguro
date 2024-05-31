package semgrep

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"

	"github.com/secguro/secguro-cli/pkg/functional"
	"github.com/secguro/secguro-cli/pkg/git"
	"github.com/secguro/secguro-cli/pkg/types"
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
	Lines    string
	Message  string
	Severity string
}

func convertSemgrepFindingToUnifiedFinding(directoryToScan string, gitMode bool,
	semgrepFinding SemgrepFinding) (types.UnifiedFinding, error) {
	gitInfo, err := git.GetGitInfo(directoryToScan, gitMode,
		"", semgrepFinding.Path, semgrepFinding.Start.Line, false)
	if err != nil {
		return types.UnifiedFinding{}, err
	}

	unifiedFinding := types.UnifiedFinding{
		Detector:    "semgrep",
		Rule:        semgrepFinding.Check_id,
		File:        "/" + semgrepFinding.Path,
		LineStart:   semgrepFinding.Start.Line,
		LineEnd:     semgrepFinding.End.Line,
		ColumnStart: semgrepFinding.Start.Col,
		ColumnEnd:   semgrepFinding.End.Col,
		Match:       semgrepFinding.Extra.Lines,
		Hint:        semgrepFinding.Extra.Message,
		Severity:    semgrepFinding.Extra.Severity,
		GitInfo:     gitInfo,
	}

	return unifiedFinding, nil
}

func getSemgrepOutputJson(directoryToScan string) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)
	semgrepOutputJsonPath := tmpDir + "/semgrepOutput.json"

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

func GetSemgrepFindingsAsUnified(directoryToScan string, gitMode bool) ([]types.UnifiedFinding, error) {
	semgrepOutputJson, err := getSemgrepOutputJson(directoryToScan)
	if err != nil {
		return nil, err
	}

	var metaSemgrepFindings Meta_SemgrepFinding
	err = json.Unmarshal(semgrepOutputJson, &metaSemgrepFindings)
	if err != nil {
		return nil, err
	}

	semgrepFindings := metaSemgrepFindings.Results

	unifiedFindings, err := functional.MapWithError(semgrepFindings,
		func(semgrepFinding SemgrepFinding) (types.UnifiedFinding, error) {
			return convertSemgrepFindingToUnifiedFinding(directoryToScan, gitMode, semgrepFinding)
		})
	if err != nil {
		return make([]types.UnifiedFinding, 0), err
	}

	return unifiedFindings, nil
}
