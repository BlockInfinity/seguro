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

	r := ""

	for i, unifiedFinding := range unifiedFindings {
		r += fmt.Sprintf("Finding %d:\n", i+1)
		r += fmt.Sprintf("  detector: %v\n", unifiedFinding.Detector)
		r += fmt.Sprintf("  rule: %v\n", unifiedFinding.Rule)
		r += fmt.Sprintf("  file: %v\n", unifiedFinding.File)
		r += fmt.Sprintf("  line: %d\n", unifiedFinding.LineStart)
		r += fmt.Sprintf("  column: %d\n", unifiedFinding.ColumnStart)
		r += fmt.Sprintf("  match: %v\n", unifiedFinding.Match)
		if len(unifiedFinding.Hint) > 0 {
			r += fmt.Sprintf("  hint: %v\n", unifiedFinding.Hint)
		}
		if gitMode {
			r += fmt.Sprintf("  commit hash: %v\n", unifiedFinding.CommitHash)
			r += fmt.Sprintf("  commit date: %v\n", unifiedFinding.CommitDate)
			r += fmt.Sprintf("  author: %v\n", unifiedFinding.AuthorName)
			r += fmt.Sprintf("  author email address: %v\n", unifiedFinding.AuthorEmailAddress)
			r += fmt.Sprintf("  commit summary: %v\n", unifiedFinding.CommitSummary)
		}
		r += "\n"
	}

	return r
}
