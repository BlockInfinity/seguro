package output

import (
	"encoding/json"
	"fmt"

	"secguro.com/secguro/pkg/functional"
	"secguro.com/secguro/pkg/types"
)

type UnifiedFindingSansGitInfo struct {
	Detector    string
	Rule        string
	File        string
	LineStart   int
	LineEnd     int
	ColumnStart int
	ColumnEnd   int
	Match       string
	Hint        string
}

func PrintJson(unifiedFindings []types.UnifiedFinding, gitMode bool) (string, error) {
	if gitMode {
		return printJsonInternal(unifiedFindings)
	} else {
		unifiedFindingsSansGitInfo := functional.Map(unifiedFindings,
			func(unifiedFinding types.UnifiedFinding) UnifiedFindingSansGitInfo {
				return UnifiedFindingSansGitInfo{
					unifiedFinding.Detector,
					unifiedFinding.Rule,
					unifiedFinding.File,
					unifiedFinding.LineStart,
					unifiedFinding.LineEnd,
					unifiedFinding.ColumnStart,
					unifiedFinding.ColumnEnd,
					unifiedFinding.Match,
					unifiedFinding.Hint,
				}
			})

		return printJsonInternal(unifiedFindingsSansGitInfo)
	}
}

func printJsonInternal[T types.UnifiedFinding | UnifiedFindingSansGitInfo](unifiedFindings []T) (string, error) {
	// Handle case of un-initialzed array (would cause
	// conversion to "null" instead of "[]").
	if unifiedFindings == nil {
		return "[]", nil
	}

	resultJson, err := json.Marshal(unifiedFindings)
	if err != nil {
		return "error", err
	}

	return string(resultJson), nil
}

func PrintText(unifiedFindings []types.UnifiedFinding, gitMode bool) string {
	if len(unifiedFindings) == 0 {
		return "no findings"
	}

	result := ""

	for i, unifiedFinding := range unifiedFindings {
		result += GetFindingTitle(i) + "\n"
		result += GetFindingBody(gitMode, unifiedFinding) + "\n"
	}

	return result
}

func GetFindingTitle(index int) string {
	return fmt.Sprintf("Finding %d:", index+1)
}

func GetFindingBody(gitMode bool, unifiedFinding types.UnifiedFinding) string {
	result := ""
	result += fmt.Sprintf("  detector: %v\n", unifiedFinding.Detector)
	result += fmt.Sprintf("  rule: %v\n", unifiedFinding.Rule)
	result += fmt.Sprintf("  match: %v\n", unifiedFinding.Match)
	result += "  location: " +
		getLocation(unifiedFinding.File, unifiedFinding.LineStart, unifiedFinding.ColumnStart)
	if gitMode && unifiedFinding.GitInfo != nil {
		result += "  historical location: " +
			getLocation(unifiedFinding.GitInfo.File, unifiedFinding.GitInfo.Line, unifiedFinding.ColumnStart)
	}
	if len(unifiedFinding.Hint) > 0 {
		result += fmt.Sprintf("  hint: %v\n", unifiedFinding.Hint)
	}
	if gitMode && unifiedFinding.GitInfo != nil {
		result += fmt.Sprintf("  commit hash: %v\n", unifiedFinding.GitInfo.CommitHash)
		result += fmt.Sprintf("  commit date: %v\n", unifiedFinding.GitInfo.CommitDate)
		result += fmt.Sprintf("  author: %v\n", unifiedFinding.GitInfo.AuthorName)
		result += fmt.Sprintf("  author email address: %v\n", unifiedFinding.GitInfo.AuthorEmailAddress)
		result += fmt.Sprintf("  commit summary: %v\n", unifiedFinding.GitInfo.CommitSummary)
	}

	return result
}

func getLocation(path string, line int, column int) string {
	if path == "" {
		return "\033[3m(does not exist)\033[0m\n"
	}

	if line != -1 && column != -1 {
		return fmt.Sprintf("%v:%v:%v\n", path, line, column)
	} else {
		return fmt.Sprintf("%v\n", path)
	}
}
