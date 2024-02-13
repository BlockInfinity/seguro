package main

import (
	"bufio"
	"os"
	"regexp"
)

func GetNumbersOfMatchingLines(filePath string, pattern string) ([]int, error) {
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
