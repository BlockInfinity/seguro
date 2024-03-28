package ignoring

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"secguro.com/secguro/pkg/config"
	"secguro.com/secguro/pkg/functional"
	"secguro.com/secguro/pkg/types"
)

const IgnoreFileName = ".secguroignore"
const SecretsIgnoreFileName = IgnoreFileName + "-secrets"

type IgnoreInstruction struct {
	FilePath   string
	LineNumber int      // -1 signifies ignoring all lines
	Rules      []string // empty array signifies ignoring all rules
}

func GetLineBasedIgnoreInstructions(unifiedFindings []types.UnifiedFinding) []IgnoreInstruction {
	filePathsWithResults := make([]string, 0)
	for _, unifiedFinding := range unifiedFindings {
		if functional.ArrayIncludes(filePathsWithResults, unifiedFinding.File) {
			continue
		}

		filePathsWithResults = append(filePathsWithResults, unifiedFinding.File)
	}

	ignoreInstructions := make([]IgnoreInstruction, 0)
	for _, filePath := range filePathsWithResults {
		lineNumbers, err := getNumbersOfMatchingLines(config.DirectoryToScan+"/"+filePath,
			"secguro-ignore-next-line")
		if err != nil {
			// Ignore failing file reads because this happens in git mode if the file has been deleted.
			continue
		}

		for _, lineNumber := range lineNumbers {
			ignoreInstructions = append(ignoreInstructions, IgnoreInstruction{
				FilePath:   filePath,
				LineNumber: lineNumber + 1,
				Rules:      make([]string, 0),
			})
		}
	}

	return ignoreInstructions
}

func getNumbersOfMatchingLines(filePath string, pattern string) ([]int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	matchingLines := make([]int, 0)
	scanner := bufio.NewScanner(file)
	lineNumber := 1

	// Compile the regular expression pattern
	regex := regexp.MustCompile(pattern)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for matches in the line
		if regex.MatchString(line) {
			matchingLines = append(matchingLines, lineNumber)
		}

		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return matchingLines, nil
}

func GetFileBasedIgnoreInstructions() ([]IgnoreInstruction, error) {
	ignoreInstructions := make([]IgnoreInstruction, 0)

	file, err := os.Open(config.DirectoryToScan + "/" + IgnoreFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return ignoreInstructions, nil
		}

		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inNewParagraph := true

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		switch {
		case strings.HasPrefix(line, "#"):
			// do nothing
		case line == "":
			inNewParagraph = true
		case inNewParagraph:
			ignoreInstructions = append(ignoreInstructions, IgnoreInstruction{
				FilePath:   line,
				LineNumber: -1,
				Rules:      make([]string, 0),
			})

			inNewParagraph = false
		default:
			ignoreInstruction := &ignoreInstructions[len(ignoreInstructions)-1]
			ignoreInstruction.Rules = append(ignoreInstruction.Rules, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ignoreInstructions, nil
}

func GetIgnoredSecrets() ([]string, error) {
	ignoredSecrets := make([]string, 0)

	file, err := os.Open(config.DirectoryToScan + "/" + SecretsIgnoreFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return ignoredSecrets, nil
		}

		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		switch {
		case strings.HasPrefix(line, "#") || line == "":
			// do nothing
		default:
			ignoredSecrets = append(ignoredSecrets, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ignoredSecrets, nil
}
