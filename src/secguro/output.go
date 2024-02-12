package main

import (
	"encoding/json"
	"fmt"
)

func printJson(unifiedFindings []UnifiedFinding) error {
	resultJson, err := json.Marshal(unifiedFindings)
	if err != nil {
		return err
	}

	fmt.Println(string(resultJson[:]))
	return nil
}

func printText(unifiedFindings []UnifiedFinding) {
	for i, unifiedFinding := range unifiedFindings {
		fmt.Printf("Finding %d:\n", i)
		fmt.Printf("  file: %v\n", unifiedFinding.File)
		fmt.Printf("  line: %d\n", unifiedFinding.Line)
		fmt.Printf("  column: %d\n", unifiedFinding.Column)
		fmt.Printf("  match: %v\n", unifiedFinding.Match)
		fmt.Printf("\n")
	}
}
