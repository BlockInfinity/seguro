package main

import (
	"encoding/json"
	"fmt"
)

func printJson(unifiedFindings []UnifiedFinding) (string, error) {
	// Handle case of un-initialzed array (would cause
	// conversion to "null" instead of "[]").
	if unifiedFindings == nil {
		return "[]", nil
	}

	resultJson, err := json.Marshal(unifiedFindings)
	if err != nil {
		return "error", err
	}

	return string(resultJson[:]), nil
}

func printText(unifiedFindings []UnifiedFinding) string {
	if len(unifiedFindings) == 0 {
		return "no findings"
	}

	r := ""

	for i, unifiedFinding := range unifiedFindings {
		r += fmt.Sprintf("Finding %d:\n", i+1)
		r += fmt.Sprintf("  detector: %v\n", unifiedFinding.Detector)
		r += fmt.Sprintf("  rule: %v\n", unifiedFinding.Rule)
		r += fmt.Sprintf("  file: %v\n", unifiedFinding.File)
		r += fmt.Sprintf("  line: %d\n", unifiedFinding.Line)
		r += fmt.Sprintf("  column: %d\n", unifiedFinding.Column)
		r += fmt.Sprintf("  match: %v\n", unifiedFinding.Match)
		if len(unifiedFinding.Hint) > 0 {
			r += fmt.Sprintf("  hint: %v\n", unifiedFinding.Hint)
		}
		r += "\n"
	}

	return r
}
