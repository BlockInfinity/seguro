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
}

// TODO: replace panic
func commandScan(scanGitHistory bool, printAsJson bool) {
	fmt.Println("Downloading and extracting dependencies...")
	err := downloadAndExtractGitleaks()
	if err != nil {
		panic(err)
	}

	fmt.Println("Scanning...")
	unifiedFindings, err := getGitleaksFindingsAsUnified()
	if err != nil {
		panic(err)
	}

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
