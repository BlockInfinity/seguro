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

type IgnoreInstruction struct {
	FilePath   string
	LineNumber int      // -1 signifies ignoring all lines
	Rules      []string // empty array signifies ignoring all rules
}

func commandScan(gitMode bool, printAsJson bool, outputDestination string, tolerance int) error {
	fmt.Println("Downloading and extracting dependencies...")
	err := downloadAndExtractGitleaks()
	if err != nil {
		return err
	}

	err = installSemgrep()
	if err != nil {
		return err
	}

	fmt.Println("Scanning...")
	unifiedFindingsGitleaks, err := getGitleaksFindingsAsUnified(gitMode)
	if err != nil {
		return err
	}

	unifiedFindingsSemgrep, err := getSemgrepFindingsAsUnified(gitMode)
	if err != nil {
		return err
	}

	unifiedFindings := []UnifiedFinding{}
	unifiedFindings = append(unifiedFindings, unifiedFindingsGitleaks...)
	unifiedFindings = append(unifiedFindings, unifiedFindingsSemgrep...)

	filePathsWithResults := set.New[string](10)
	for _, unifiedFinding := range unifiedFindings {
		filePathsWithResults.Insert(unifiedFinding.File)
	}

	ignoreInstructions := make([]IgnoreInstruction, 10)
	filePathsWithResults.ForEach(func(filePath string) bool {
		lineNumbers, err := GetNumbersOfMatchingLines(directoryToScan+"/"+filePath, "secguro-ignore-next-line")
		if err != nil {
			// Ignore failing file reads because this happens in git mode if the file has been deleted.
			return false
		}

		for _, lineNumber := range lineNumbers {
			ignoreInstructions = append(ignoreInstructions, IgnoreInstruction{
				FilePath:   filePath,
				LineNumber: lineNumber + 1,
				Rules:      make([]string, 0),
			})
		}

		return false
	})

	unifiedFindingsNotIgnored := Filter(unifiedFindings, func(unifiedFinding UnifiedFinding) bool {
		for _, ignoreInstruction := range ignoreInstructions {
			if ignoreInstruction.FilePath == unifiedFinding.File &&
				(ignoreInstruction.LineNumber == unifiedFinding.LineStart || unifiedFinding.LineStart == -1) &&
				(len(ignoreInstruction.Rules) == 0 ||
					arrayIncludes(ignoreInstruction.Rules, unifiedFinding.Rule)) {
				return false
			}
		}
		return true
	})

	var output string
	if printAsJson {
		output, err = printJson(unifiedFindingsNotIgnored, gitMode)
		if err != nil {
			return err
		}
	} else {
		output = printText(unifiedFindingsNotIgnored, gitMode)
	}

	if outputDestination == "" {
		fmt.Println("Findings:")
		fmt.Println(output)
	} else {
		err = os.WriteFile(outputDestination, []byte(output), 0644)
		if err != nil {
			return err
		}

		fmt.Println("Output written to: " + outputDestination)
	}

	exitWithAppropriateExitCode(len(unifiedFindingsNotIgnored), tolerance)
	return nil
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
