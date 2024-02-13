package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-set/v2"
)

const maxFindingsIndicatingExitCode = 250

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

type FilePathWithLineNumber struct {
	FilePath   string
	LineNumber int
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

	filePathsWithResults := set.New[string](10)
	for _, unifiedFinding := range unifiedFindings {
		filePathsWithResults.Insert(unifiedFinding.File)
	}

	ignoredLines := set.New[FilePathWithLineNumber](10)
	filePathsWithResults.ForEach(func(filePath string) bool {
		lineNumbers, err := GetNumbersOfMatchingLines(filePath, "secguro-ignore-next-line")
		if err != nil {
			panic(err)
		}

		for _, lineNumber := range lineNumbers {
			ignoredLines.Insert(FilePathWithLineNumber{
				FilePath:   filePath,
				LineNumber: lineNumber + 1,
			})
		}

		return false
	})

	unifiedFindingsNotIgnored := Filter(unifiedFindings, func(unifiedFinding UnifiedFinding) bool {
		r := true
		ignoredLines.ForEach(func(ignoredLine FilePathWithLineNumber) bool {
			if ignoredLine.FilePath == unifiedFinding.File && ignoredLine.LineNumber == unifiedFinding.Line {
				r = false
				return true
			}

			return false
		})
		return r
	})

	fmt.Println("Findings:")
	if printAsJson {
		err = printJson(unifiedFindingsNotIgnored)
		if err != nil {
			panic(err)
		}
	} else {
		printText(unifiedFindingsNotIgnored)
	}

	exitCode := len(unifiedFindingsNotIgnored)
	if exitCode > maxFindingsIndicatingExitCode {
		exitCode = maxFindingsIndicatingExitCode
	}
	os.Exit(exitCode)
}
