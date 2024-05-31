package scan

import (
	"fmt"
	"os"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
	"github.com/secguro/secguro-cli/pkg/dependencies"
	"github.com/secguro/secguro-cli/pkg/dependencycheck"
	"github.com/secguro/secguro-cli/pkg/functional"
	"github.com/secguro/secguro-cli/pkg/gitleaks"
	"github.com/secguro/secguro-cli/pkg/ignoring"
	"github.com/secguro/secguro-cli/pkg/output"
	"github.com/secguro/secguro-cli/pkg/reporting"
	"github.com/secguro/secguro-cli/pkg/semgrep"
	"github.com/secguro/secguro-cli/pkg/types"
)

const maxFindingsIndicatingExitCode = 250

func CommandScan(directoryToScan string, gitMode bool, disabledDetectors []string,
	printAsJson bool, outputDestination string, tolerance int) error {
	unifiedFindingsNotIgnored, failedDetectors, err := PerformScan(directoryToScan, gitMode, disabledDetectors)
	if err != nil {
		return err
	}

	err = writeOutput(gitMode, printAsJson, outputDestination, unifiedFindingsNotIgnored)
	if err != nil {
		return err
	}

	err = reporting.ReportScanIfApplicable(directoryToScan, unifiedFindingsNotIgnored)
	if err != nil {
		return err
	}

	if len(failedDetectors) != 0 {
		fmt.Println("Be mindful that some detectors have failed. Confer top of output.")
	}

	exitWithAppropriateExitCode(len(unifiedFindingsNotIgnored), tolerance)

	return nil
}

func PerformScan(directoryToScan string,
	gitMode bool, disabledDetectors []string) ([]types.UnifiedFinding, []string, error) {
	fmt.Print("Downloading and extracting dependencies...")
	err := dependencies.InstallDependencies(disabledDetectors)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("done")

	fmt.Print("Scanning...")
	unifiedFindings, failedDetectors := getUnifiedFindings(directoryToScan, gitMode, disabledDetectors)
	if len(failedDetectors) == 0 {
		fmt.Println("done")
	} else {
		fmt.Println("done with errors: the following detectors failed:")
		for _, failedDetector := range failedDetectors {
			fmt.Println("  • " + failedDetector)
		}
	}

	unifiedFindingsNotIgnored, err := getFindingsNotIgnored(directoryToScan, unifiedFindings)
	if err != nil {
		return nil, nil, err
	}

	return unifiedFindingsNotIgnored, failedDetectors, nil
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

func getUnifiedFindings(directoryToScan string,
	gitMode bool, disabledDetectors []string) ([]types.UnifiedFinding, []string) {
	failedDetectors := make([]string, 0)
	unifiedFindings := make([]types.UnifiedFinding, 0)

	if !functional.ArrayIncludes(disabledDetectors, "gitleaks") {
		unifiedFindingsGitleaks, err := gitleaks.GetGitleaksFindingsAsUnified(directoryToScan, gitMode)
		if err != nil {
			failedDetectors = append(failedDetectors, "gitleaks")
		} else {
			unifiedFindings = append(unifiedFindings, unifiedFindingsGitleaks...)
		}
	}

	if !functional.ArrayIncludes(disabledDetectors, "semgrep") {
		unifiedFindingsSemgrep, err := semgrep.GetSemgrepFindingsAsUnified(directoryToScan, gitMode)
		if err != nil {
			failedDetectors = append(failedDetectors, "semgrep")
		} else {
			unifiedFindings = append(unifiedFindings, unifiedFindingsSemgrep...)
		}
	}

	if !functional.ArrayIncludes(disabledDetectors, "dependencycheck") {
		unifiedFindingsDependencycheck, err :=
			dependencycheck.GetDependencycheckFindingsAsUnified(directoryToScan, gitMode)
		if err != nil {
			failedDetectors = append(failedDetectors, "dependencycheck")
		} else {
			unifiedFindings = append(unifiedFindings, unifiedFindingsDependencycheck...)
		}
	}

	return unifiedFindings, failedDetectors
}

func getFindingsNotIgnored(directoryToScan string, //nolint: cyclop
	unifiedFindings []types.UnifiedFinding) ([]types.UnifiedFinding, error) {
	lineBasedIgnoreInstructions := ignoring.GetLineBasedIgnoreInstructions(directoryToScan, unifiedFindings)
	fileBasedIgnoreInstructions, err := ignoring.GetFileBasedIgnoreInstructions(directoryToScan)
	if err != nil {
		return make([]types.UnifiedFinding, 0), err
	}

	ignoreInstructions := []ignoring.IgnoreInstruction{
		// Ignore .secguroignore and .secguroignore-secrets in case
		// a detector finds something in there in the future (does
		// not currently appear to be the case).
		{
			FilePath:   "/" + ignoring.IgnoreFileName,
			LineNumber: -1,
			Rules:      make([]string, 0),
		},
		{
			FilePath:   "/" + ignoring.SecretsIgnoreFileName,
			LineNumber: -1,
			Rules:      make([]string, 0),
		},
	}
	ignoreInstructions = append(ignoreInstructions, lineBasedIgnoreInstructions...)
	ignoreInstructions = append(ignoreInstructions, fileBasedIgnoreInstructions...)

	ignoredSecrets, err := ignoring.GetIgnoredSecrets(directoryToScan)
	if err != nil {
		return make([]types.UnifiedFinding, 0), err
	}

	unifiedFindingsNotIgnored := functional.Filter(unifiedFindings, func(unifiedFinding types.UnifiedFinding) bool {
		// Filter findings based on rules ignored for specific paths as well as on specific lines.
		for _, ii := range ignoreInstructions {
			gitIgnoreMatcher := ignore.CompileIgnoreLines(ii.FilePath)
			if gitIgnoreMatcher.MatchesPath(unifiedFinding.File) &&
				(ii.LineNumber == unifiedFinding.LineStart || ii.LineNumber == -1) &&
				(len(ii.Rules) == 0 || functional.ArrayIncludes(ii.Rules, unifiedFinding.Rule)) {
				return false
			}
		}

		// Filter findings based on ignored secrets
		for _, ignoredSecret := range ignoredSecrets {
			if !IsSecretDetectionRule(unifiedFinding.Rule) {
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
	outputDestination string, unifiedFindingsNotIgnored []types.UnifiedFinding) error {
	var outputString string
	if printAsJson {
		var err error
		outputString, err = output.PrintJson(unifiedFindingsNotIgnored, gitMode)
		if err != nil {
			return err
		}
	} else {
		outputString = output.PrintText(unifiedFindingsNotIgnored, gitMode)
	}

	if outputDestination == "" {
		fmt.Println("Findings:")
		fmt.Println(outputString)
	} else {
		const filePermissions = 0644
		err := os.WriteFile(outputDestination, []byte(outputString), filePermissions)
		if err != nil {
			return err
		}

		fmt.Println("Output written to: " + outputDestination)
	}

	return nil
}
