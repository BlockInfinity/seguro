package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-set/v2"
)

const maxFindingsIndicatingExitCode = 250

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

type GitInfo struct {
	CommitHash         string
	CommitDate         string
	AuthorName         string
	AuthorEmailAddress string
	CommitSummary      string
}

// The attributes need to start with capital letter because
// otherwise the JSON formatter cannot see them.
type UnifiedFinding struct {
	Detector           string
	Rule               string
	File               string
	LineStart          int
	LineEnd            int
	ColumnStart        int
	ColumnEnd          int
	Match              string
	Hint               string
	CommitHash         string
	CommitDate         string
	AuthorName         string
	AuthorEmailAddress string
	CommitSummary      string
}

type FilePathWithLineNumber struct {
	FilePath   string
	LineNumber int
}

// TODO: replace panic
func commandScan(gitMode bool, printAsJson bool, outputDestination string, tolerance int) {
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
	unifiedFindingsGitleaks, err := getGitleaksFindingsAsUnified(gitMode)
	if err != nil {
		panic(err)
	}

	unifiedFindingsSemgrep, err := getSemgrepFindingsAsUnified(gitMode)
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
		lineNumbers, err := GetNumbersOfMatchingLines(directoryToScan+"/"+filePath, "secguro-ignore-next-line")
		if err != nil {
			// Ignore failing file reads because this happens in git mode if the file has been deleted.
			return false
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
			if ignoredLine.FilePath == unifiedFinding.File &&
				ignoredLine.LineNumber == unifiedFinding.LineStart {
				r = false
			}

			// It should be possible to end the forEach early as soon as r
			// is set to false. However, this causes undeterministically
			// occurring behavior that causes subsequent matches of ignored
			// lines to be missed.
			return true
		})
		return r
	})

	output := (func() string {
		if printAsJson {
			o, err := printJson(unifiedFindingsNotIgnored, gitMode)
			if err != nil {
				panic(err)
			}
			return o
		} else {
			return printText(unifiedFindingsNotIgnored, gitMode)
		}
	})()

	if outputDestination == "" {
		fmt.Println("Findings:")
		fmt.Println(output)
	} else {
		err = os.WriteFile(outputDestination, []byte(output), 0644)
		if err != nil {
			panic(err)
		}

		fmt.Println("Output written to: " + outputDestination)
	}

	exitWithAppropriateExitCode(len(unifiedFindingsNotIgnored), tolerance)
}

func exitWithAppropriateExitCode(numberOfFindingsNotIgnored int, tolerance int) {
	if numberOfFindingsNotIgnored <= tolerance {
		os.Exit(0)
	}

	if numberOfFindingsNotIgnored > maxFindingsIndicatingExitCode {
		os.Exit(maxFindingsIndicatingExitCode)
	}

	os.Exit(numberOfFindingsNotIgnored)
}
