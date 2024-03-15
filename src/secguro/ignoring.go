package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

type IgnoreInstruction struct {
	FilePath   string
	LineNumber int      // -1 signifies ignoring all lines
	Rules      []string // empty array signifies ignoring all rules
}

func getLineBasedIgnoreInstructions(unifiedFindings []UnifiedFinding) []IgnoreInstruction {
	filePathsWithResults := make([]string, 0)
	for _, unifiedFinding := range unifiedFindings {
		if arrayIncludes(filePathsWithResults, unifiedFinding.File) {
			continue
		}

		filePathsWithResults = append(filePathsWithResults, unifiedFinding.File)
	}

	ignoreInstructions := make([]IgnoreInstruction, 0)
	for _, filePath := range filePathsWithResults {
		lineNumbers, err := getNumbersOfMatchingLines(directoryToScan+"/"+filePath, "secguro-ignore-next-line")
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

func getFileBasedIgnoreInstructions() ([]IgnoreInstruction, error) {
	ignoreInstructions := make([]IgnoreInstruction, 0)

	file, err := os.Open(directoryToScan + "/.secguroignore")
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

		if strings.HasPrefix(line, "#") {
			continue
		}

		if line == "" {
			inNewParagraph = true
			continue
		}

		if inNewParagraph {
			ignoreInstructions = append(ignoreInstructions, IgnoreInstruction{
				FilePath:   line,
				LineNumber: -1,
				Rules:      make([]string, 0),
			})

			inNewParagraph = false
		} else {
			ignoreInstruction := &ignoreInstructions[len(ignoreInstructions)-1]
			ignoreInstruction.Rules = append(ignoreInstruction.Rules, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ignoreInstructions, nil
}

func getIgnoredSecrets() ([]string, error) {
	ignoredSecrets := make([]string, 0)

	file, err := os.Open(directoryToScan + "/.secguroignore-secrets")
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

		if strings.HasPrefix(line, "#") {
			continue
		}

		if line == "" {
			continue
		}

		ignoredSecrets = append(ignoredSecrets, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ignoredSecrets, nil
}
