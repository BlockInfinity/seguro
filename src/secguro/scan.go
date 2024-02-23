package main

import (
	"fmt"
	"os"

	ignore "github.com/sabhiram/go-gitignore"
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

	lineBasedIgnoreInstructions := getLineBasedIgnoreInstructions(unifiedFindings)
	fileBasedIgnoreInstructions, err := getFileBasedIgnoreInstructions()
	if err != nil {
		return err
	}

	ignoreInstructions := []IgnoreInstruction{}
	ignoreInstructions = append(ignoreInstructions, lineBasedIgnoreInstructions...)
	ignoreInstructions = append(ignoreInstructions, fileBasedIgnoreInstructions...)

	unifiedFindingsNotIgnored := Filter(unifiedFindings, func(unifiedFinding UnifiedFinding) bool {
		for _, ii := range ignoreInstructions {
			gitIgnoreMatcher := ignore.CompileIgnoreLines(ii.FilePath)
			if gitIgnoreMatcher.MatchesPath(unifiedFinding.File) &&
				(ii.LineNumber == unifiedFinding.LineStart || ii.LineNumber == -1) &&
				(len(ii.Rules) == 0 || arrayIncludes(ii.Rules, unifiedFinding.Rule)) {
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
