package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

/**
 * Returns git info if in git mode; otherwise returns nil.
 */
func getGitInfo(gitMode bool, revision string, filePath string, lineNumber int) (*GitInfo, error) {
	if !gitMode {
		return nil, nil //nolint: nilnil
	}

	gitBlameOutput, err := getGitBlameOutput(revision, filePath, lineNumber)

	// If the file is not tracked with git, getGitBlameOutput() returns an error
	// because `git blame` exits with exit code 128. However, this behavior does
	// not seem to be documented. Therefore, the exit code is not checked here.
	// Instead, failure of `git blame` is assumed to always mean that the file
	// is not tracked with git.
	if err != nil {
		return nil, nil //nolint: nilnil, nilerr
	}

	gitInfo, err := parseGitBlameOutput(gitBlameOutput)

	return &gitInfo, err
}

/**
 * Empty string for revision means working directory.
 */
func getGitBlameOutput(revision string, filePath string, lineNumber int) ([]byte, error) {
	lineRange := fmt.Sprintf("%d,%d", lineNumber, lineNumber)
	cmd := (func() *exec.Cmd {
		if revision == "" {
			return exec.Command("git", "blame", "-L", lineRange, "-p", "--", filePath)
		} else {
			return exec.Command("git", "blame", "-L", lineRange, "-p", revision, "--", filePath)
		}
	})()

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
			continue
		}

		//nolint: gocritic
		if strings.HasPrefix(line, "summary ") {
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
		// Ignore this error because it can only happen with the last line of the output
		// (the line git blame was called on) as the others are always short enough.
		// This line is not used by this function.
		if err.Error() != "bufio.Scanner: token too long" {
			return GitInfo{}, err
		}
	}

	return gitInfo, nil
}
