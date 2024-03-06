package main

import (
	"encoding/json"
	"fmt"
)

func printJson(unifiedFindings []UnifiedFinding, gitMode bool) (string, error) {
	if gitMode {
		return printJsonInternal(unifiedFindings)
	} else {
		unifiedFindingsSansGitInfo := Map(unifiedFindings, func(unifiedFinding UnifiedFinding) UnifiedFindingSansGitInfo {
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

func printJsonInternal[T UnifiedFinding | UnifiedFindingSansGitInfo](unifiedFindings []T) (string, error) {
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

func printText(unifiedFindings []UnifiedFinding, gitMode bool) string {
	if len(unifiedFindings) == 0 {
		return "no findings"
	}

	result := ""

	for i, unifiedFinding := range unifiedFindings {
		result += fmt.Sprintf("Finding %d:\n", i+1)
		result += fmt.Sprintf("  detector: %v\n", unifiedFinding.Detector)
		result += fmt.Sprintf("  rule: %v\n", unifiedFinding.Rule)
		if unifiedFinding.LineStart != -1 && unifiedFinding.ColumnStart != -1 {
			result += fmt.Sprintf("  location: %v:%v:%v\n",
				unifiedFinding.File, unifiedFinding.LineStart, unifiedFinding.ColumnStart)
		} else {
			result += fmt.Sprintf("  location: %v\n", unifiedFinding.File)
		}
		result += fmt.Sprintf("  match: %v\n", unifiedFinding.Match)
		if len(unifiedFinding.Hint) > 0 {
			result += fmt.Sprintf("  hint: %v\n", unifiedFinding.Hint)
		}
		if gitMode {
			result += fmt.Sprintf("  commit hash: %v\n", unifiedFinding.CommitHash)
			result += fmt.Sprintf("  commit date: %v\n", unifiedFinding.CommitDate)
			result += fmt.Sprintf("  author: %v\n", unifiedFinding.AuthorName)
			result += fmt.Sprintf("  author email address: %v\n", unifiedFinding.AuthorEmailAddress)
			result += fmt.Sprintf("  commit summary: %v\n", unifiedFinding.CommitSummary)
		}
		result += "\n"
	}

	return result
}
