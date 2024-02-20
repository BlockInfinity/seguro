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
		return GitInfo{}, nil // nolint: exhaustruct
	}

	lineRange := fmt.Sprintf("%d,%d", lineNumber, lineNumber)
	cmd := exec.Command("git", "blame", "-L", lineRange, "-p", filePath)
	cmd.Dir = directoryToScan
	out, err := cmd.Output()
	if err != nil {
		return GitInfo{}, err // nolint: exhaustruct
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	r := GitInfo{} // nolint: exhaustruct
	isFirstLine := true
	for scanner.Scan() {
		line := scanner.Text()

		if isFirstLine {
			r.CommitHash = strings.Fields(line)[0]
			isFirstLine = false
		} else if strings.HasPrefix(line, "summary ") {
			r.CommitMessage = strings.TrimPrefix(line, "summary ")
		} else if strings.HasPrefix(line, "author ") {
			r.AuthorName = strings.TrimPrefix(line, "author ")
		} else if strings.HasPrefix(line, "author-mail ") {
			r.AuthorEmailAddress = strings.TrimSuffix(strings.TrimPrefix(line, "author-mail <"), ">")
		} else if strings.HasPrefix(line, "author-time ") {
			authorTimeString := strings.TrimPrefix(line, "author-time ")
			authorTimeInt, err := strconv.Atoi(authorTimeString)
			if err != nil {
				return GitInfo{}, err // nolint: exhaustruct
			}
			authorTime := time.Unix(int64(authorTimeInt), 0)
			authorTimeFormatted := authorTime.UTC().Format(time.RFC3339)
			r.CommitDate = authorTimeFormatted
		}
	}

	if err := scanner.Err(); err != nil {
		return GitInfo{}, err // nolint: exhaustruct
	}

	return r, nil
}
