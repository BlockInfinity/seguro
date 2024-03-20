package main

import (
	"fmt"
	"os"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

const maxFindingsIndicatingExitCode = 250

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
	Detector    string
	Rule        string
	File        string
	LineStart   int
	LineEnd     int
	ColumnStart int
	ColumnEnd   int
	Match       string
	Hint        string
	GitInfo     *GitInfo
}

func commandScan(gitMode bool, disabledDetectors []string,
	printAsJson bool, outputDestination string, tolerance int) error {
	unifiedFindingsNotIgnored, err := performScan(gitMode, disabledDetectors)
	if err != nil {
		return err
	}

	err = writeOutput(gitMode, printAsJson, outputDestination, unifiedFindingsNotIgnored)
	if err != nil {
		return err
	}

	exitWithAppropriateExitCode(len(unifiedFindingsNotIgnored), tolerance)

	return nil
}

func performScan(gitMode bool, disabledDetectors []string) ([]UnifiedFinding, error) {
	fmt.Print("Downloading and extracting dependencies...")
	err := installDependencies(disabledDetectors)
	if err != nil {
		return nil, err
	}
	fmt.Println("done")

	fmt.Print("Scanning...")
	unifiedFindings, err := getUnifiedFindings(gitMode, disabledDetectors)
	if err != nil {
		return nil, err
	}
	fmt.Println("done")

	unifiedFindingsNotIgnored, err := getFindingsNotIgnored(unifiedFindings)
	if err != nil {
		return nil, err
	}

	return unifiedFindingsNotIgnored, nil
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

func installDependencies(disabledDetectors []string) error {
	if !arrayIncludes(disabledDetectors, "gitleaks") {
		err := downloadAndExtractGitleaks()
		if err != nil {
			return err
		}
	}

	if !arrayIncludes(disabledDetectors, "semgrep") {
		err := installSemgrep()
		if err != nil {
			return err
		}
	}

	if !arrayIncludes(disabledDetectors, "dependencycheck") {
		err := downloadAndExtractDependencycheck()
		if err != nil {
			return err
		}
	}

	return nil
}

func getUnifiedFindings(gitMode bool, disabledDetectors []string) ([]UnifiedFinding, error) {
	unifiedFindings := make([]UnifiedFinding, 0)

	if !arrayIncludes(disabledDetectors, "gitleaks") {
		unifiedFindingsGitleaks, err := getGitleaksFindingsAsUnified(gitMode)
		if err != nil {
			return unifiedFindings, err
		}
		unifiedFindings = append(unifiedFindings, unifiedFindingsGitleaks...)
	}

	if !arrayIncludes(disabledDetectors, "semgrep") {
		unifiedFindingsSemgrep, err := getSemgrepFindingsAsUnified(gitMode)
		if err != nil {
			return unifiedFindings, err
		}
		unifiedFindings = append(unifiedFindings, unifiedFindingsSemgrep...)
	}

	if !arrayIncludes(disabledDetectors, "dependencycheck") {
		unifiedFindingsDependencycheck, err := getDependencycheckFindingsAsUnified(gitMode)
		if err != nil {
			return unifiedFindings, err
		}
		unifiedFindings = append(unifiedFindings, unifiedFindingsDependencycheck...)
	}

	return unifiedFindings, nil
}

func getFindingsNotIgnored(unifiedFindings []UnifiedFinding) ([]UnifiedFinding, error) { //nolint: cyclop
	lineBasedIgnoreInstructions := getLineBasedIgnoreInstructions(unifiedFindings)
	fileBasedIgnoreInstructions, err := getFileBasedIgnoreInstructions()
	if err != nil {
		return make([]UnifiedFinding, 0), err
	}

	ignoreInstructions := []IgnoreInstruction{
		// Ignore .secguroignore and .secguroignore-secrets in case
		// a detector finds something in there in the future (does
		// not currently appear to be the case).
		{
			FilePath:   "/" + ignoreFileName,
			LineNumber: -1,
			Rules:      make([]string, 0),
		},
		{
			FilePath:   "/" + secretsIgnoreFileName,
			LineNumber: -1,
			Rules:      make([]string, 0),
		},
	}
	ignoreInstructions = append(ignoreInstructions, lineBasedIgnoreInstructions...)
	ignoreInstructions = append(ignoreInstructions, fileBasedIgnoreInstructions...)

	ignoredSecrets, err := getIgnoredSecrets()
	if err != nil {
		return make([]UnifiedFinding, 0), err
	}

	unifiedFindingsNotIgnored := Filter(unifiedFindings, func(unifiedFinding UnifiedFinding) bool {
		// Filter findings based on rules ignored for specific paths as well as on specific lines.
		for _, ii := range ignoreInstructions {
			gitIgnoreMatcher := ignore.CompileIgnoreLines(ii.FilePath)
			if gitIgnoreMatcher.MatchesPath(unifiedFinding.File) &&
				(ii.LineNumber == unifiedFinding.LineStart || ii.LineNumber == -1) &&
				(len(ii.Rules) == 0 || arrayIncludes(ii.Rules, unifiedFinding.Rule)) {
				return false
			}
		}

		// Filter findings based on ignored secrets
		for _, ignoredSecret := range ignoredSecrets {
			if !isSecretDetectionRule(unifiedFinding.Rule) {
				continue
			}

			if strings.Contains(unifiedFinding.Match, ignoredSecret) {
				return false
			}
		}

		return true
	})

	return unifiedFindingsNotIgnored, nil
}

func writeOutput(gitMode bool, printAsJson bool,
	outputDestination string, unifiedFindingsNotIgnored []UnifiedFinding) error {
	var output string
	if printAsJson {
		var err error
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
		const filePermissions = 0644
		err := os.WriteFile(outputDestination, []byte(output), filePermissions)
		if err != nil {
			return err
		}

		fmt.Println("Output written to: " + outputDestination)
	}

	return nil
}
