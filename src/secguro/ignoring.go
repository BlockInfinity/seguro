package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/go-set/v2"
)

type IgnoreInstruction struct {
	FilePath   string
	LineNumber int      // -1 signifies ignoring all lines
	Rules      []string // empty array signifies ignoring all rules
}

func getLineBasedIgnoreInstructions(unifiedFindings []UnifiedFinding) []IgnoreInstruction {
	filePathsWithResults := set.New[string](10)
	for _, unifiedFinding := range unifiedFindings {
		filePathsWithResults.Insert(unifiedFinding.File)
	}

	ignoreInstructions := make([]IgnoreInstruction, 0)
	filePathsWithResults.ForEach(func(filePath string) bool {
		lineNumbers, err := getNumbersOfMatchingLines(directoryToScan+"/"+filePath, "secguro-ignore-next-line")
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

	return ignoreInstructions
}

func getNumbersOfMatchingLines(filePath string, pattern string) ([]int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matchingLines []int
	scanner := bufio.NewScanner(file)
	lineNumber := 1

	// Compile the regular expression pattern
	re := regexp.MustCompile(pattern)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for matches in the line
		if re.MatchString(line) {
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
	// TODO: handle case of file not existing
	file, err := os.Open(".secguroignore")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ignoreInstructions []IgnoreInstruction
	scanner := bufio.NewScanner(file)
	inNewParagraph := true

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

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
			ignoreInstruction := ignoreInstructions[len(ignoreInstructions)-1]
			ignoreInstruction.Rules = append(ignoreInstruction.Rules, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ignoreInstructions, nil
}
