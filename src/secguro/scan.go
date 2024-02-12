package main

import (
	"fmt"
	"os"
)

// The attributes need to start with capital letter because
// otherwise the JSON formatter cannot see them.
type UnifiedFinding struct {
	Detector string
	Rule     string
	File     string
	Line     int
	Column   int
	Match    string
	Hint     string
}

// TODO: replace panic.
func commandScan(scanGitHistory bool, printAsJson bool) {
	fmt.Println("Downloading and extracting dependencies...")
	err := downloadAndExtractGitleaks()
	if err != nil {
		panic(err)
	}

	err = installSemgrep()
	if err != nil {
		panic(err)
	}

	fmt.Println("Scanning...")
	unifiedFindingsGitleaks, err := getGitleaksFindingsAsUnified()
	if err != nil {
		panic(err)
	}

	unifiedFindingsSemgrep, err := getSemgrepFindingsAsUnified()
	if err != nil {
		panic(err)
	}

	unifiedFindings := []UnifiedFinding{}
	unifiedFindings = append(unifiedFindings, unifiedFindingsGitleaks...)
	unifiedFindings = append(unifiedFindings, unifiedFindingsSemgrep...)

	fmt.Println("Findings:")
	if printAsJson {
		err = printJson(unifiedFindings)
		if err != nil {
			panic(err)
		}
	} else {
		printText(unifiedFindings)
	}

	os.Exit(0)
}
