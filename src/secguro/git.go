package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func getGitInfo(filePath string, lineNumber int, gitMode bool) (GitInfo, error) {
	if !gitMode {
		return GitInfo{}, nil //nolint: exhaustruct
	}

	gitBlameOutput, err := getGitBlameOutput(filePath, lineNumber)
	if err != nil {
		return GitInfo{}, err
	}

	gitInfo, err := parseGitBlameOutput(gitBlameOutput)

	return gitInfo, err
}

func getGitBlameOutput(filePath string, lineNumber int) ([]byte, error) {
	lineRange := fmt.Sprintf("%d,%d", lineNumber, lineNumber)
	cmd := exec.Command("git", "blame", "-L", lineRange, "-p", filePath)
	cmd.Dir = directoryToScan
	gitBlameOutput, err := cmd.Output()

	return gitBlameOutput, err
}

func parseGitBlameOutput(gitBlameOutput []byte) (GitInfo, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(gitBlameOutput)))
	gitInfo := GitInfo{} //nolint: exhaustruct
	isFirstLine := true
	for scanner.Scan() {
		line := scanner.Text()

		if isFirstLine {
			gitInfo.CommitHash = strings.Fields(line)[0]
			isFirstLine = false
		} else if strings.HasPrefix(line, "summary ") {
			gitInfo.CommitSummary = strings.TrimPrefix(line, "summary ")
		} else if strings.HasPrefix(line, "author ") {
			gitInfo.AuthorName = strings.TrimPrefix(line, "author ")
		} else if strings.HasPrefix(line, "author-mail ") {
			gitInfo.AuthorEmailAddress = strings.TrimSuffix(strings.TrimPrefix(line, "author-mail <"), ">")
		} else if strings.HasPrefix(line, "author-time ") {
			authorTimeString := strings.TrimPrefix(line, "author-time ")
			authorTimeInt, err := strconv.Atoi(authorTimeString)
			if err != nil {
				return GitInfo{}, err
			}
			authorTime := time.Unix(int64(authorTimeInt), 0)
			authorTimeFormatted := authorTime.UTC().Format(time.RFC3339)
			gitInfo.CommitDate = authorTimeFormatted
		}
	}

	if err := scanner.Err(); err != nil {
		return GitInfo{}, err
	}

	return gitInfo, nil
}
