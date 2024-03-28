package semgrep

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"

	"secguro.com/secguro/pkg/config"
	"secguro.com/secguro/pkg/dependencies"
	"secguro.com/secguro/pkg/functional"
	"secguro.com/secguro/pkg/git"
	"secguro.com/secguro/pkg/types"
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

func convertSemgrepFindingToUnifiedFinding(gitMode bool, semgrepFinding SemgrepFinding) (types.UnifiedFinding, error) {
	gitInfo, err := git.GetGitInfo(gitMode, "", semgrepFinding.Path, semgrepFinding.Start.Line, false)
	if err != nil {
		return types.UnifiedFinding{}, err
	}

	unifiedFinding := types.UnifiedFinding{
		Detector:    "semgrep",
		Rule:        semgrepFinding.Check_id,
		File:        semgrepFinding.Path,
		LineStart:   semgrepFinding.Start.Line,
		LineEnd:     semgrepFinding.End.Line,
		ColumnStart: semgrepFinding.Start.Col,
		ColumnEnd:   semgrepFinding.End.Col,
		Match:       semgrepFinding.Extra.Lines,
		Hint:        semgrepFinding.Extra.Message,
		GitInfo:     gitInfo,
	}

	return unifiedFinding, nil
}

func getSemgrepOutputJson() ([]byte, error) {
	semgrepOutputJsonPath := dependencies.DependenciesDir + "/semgrepOutput.json"

	cmd := exec.Command("semgrep", "scan", "--json", "-o", semgrepOutputJsonPath)
	cmd.Dir = config.DirectoryToScan
	// Ignore error because this is expected to deliver an exit code not equal to 0 and write to stderr.
	out, _ := cmd.Output()
	if len(out) != 0 {
		return nil, errors.New("received unexpected output from semgrep")
	}

	semgrepOutputJson, err := os.ReadFile(semgrepOutputJsonPath)

	return semgrepOutputJson, err
}

func GetSemgrepFindingsAsUnified(gitMode bool) ([]types.UnifiedFinding, error) {
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

	unifiedFindings, err := functional.MapWithError(semgrepFindings,
		func(semgrepFinding SemgrepFinding) (types.UnifiedFinding, error) {
			return convertSemgrepFindingToUnifiedFinding(gitMode, semgrepFinding)
		})
	if err != nil {
		return make([]types.UnifiedFinding, 0), err
	}

	return unifiedFindings, nil
}
