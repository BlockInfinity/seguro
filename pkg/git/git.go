package git

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/secguro/secguro-cli/pkg/functional"
	"github.com/secguro/secguro-cli/pkg/types"
)

/**
 * Returns git info if in git mode; otherwise returns nil.
 */
func GetGitInfo(directoryToScan string, gitMode bool, revision string,
	filePath string, lineNumber int, reverse bool) (*types.GitInfo, error) {
	if !gitMode {
		return nil, nil //nolint: nilnil
	}

	gitBlameOutput, err := getGitBlameOutput(directoryToScan, revision, filePath, lineNumber, reverse)

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
func getGitBlameOutput(directoryToScan string,
	revision string, filePath string, lineNumber int, reverse bool) ([]byte, error) {
	lineRange := fmt.Sprintf("%d,%d", lineNumber, lineNumber)

	args := []string{"blame", "-p", "-L", lineRange}
	if revision == "" { //nolint: nestif
		if reverse {
			return make([]byte, 0), errors.New(
				"git blame in reverse does not make sense without a given revision")
		} else { //nolint: staticcheck
			// args do not need to be modified
		}
	} else {
		if reverse {
			args = append(args, "--reverse", revision+"..HEAD")
		} else {
			args = append(args, revision)
		}
	}

	args = append(args, "--", filePath)

	cmd := exec.Command("git", args...)
	cmd.Dir = directoryToScan
	gitBlameOutput, err := cmd.Output()

	return gitBlameOutput, err
}

func parseGitBlameOutput(gitBlameOutput []byte) (types.GitInfo, error) { //nolint: cyclop
	scanner := bufio.NewScanner(strings.NewReader(string(gitBlameOutput)))
	gitInfo := types.GitInfo{} //nolint: exhaustruct
	isFirstLine := true
	for scanner.Scan() {
		line := scanner.Text()

		if isFirstLine {
			lineFields := strings.Fields(line)
			gitInfo.CommitHash = lineFields[0]
			lineNumber, err := strconv.Atoi(lineFields[1])
			if err != nil {
				return types.GitInfo{}, err
			}
			gitInfo.Line = lineNumber
			isFirstLine = false

			continue
		}

		switch {
		case strings.HasPrefix(line, "summary "):
			gitInfo.CommitSummary = strings.TrimPrefix(line, "summary ")
		case strings.HasPrefix(line, "author "):
			gitInfo.AuthorName = strings.TrimPrefix(line, "author ")
		case strings.HasPrefix(line, "author-mail "):
			gitInfo.AuthorEmailAddress = strings.TrimSuffix(strings.TrimPrefix(line, "author-mail <"), ">")
		case strings.HasPrefix(line, "author-time "):
			authorTimeString := strings.TrimPrefix(line, "author-time ")
			authorTimeInt, err := strconv.Atoi(authorTimeString)
			if err != nil {
				return types.GitInfo{}, err
			}
			authorTime := time.Unix(int64(authorTimeInt), 0)
			authorTimeFormatted := authorTime.UTC().Format(time.RFC3339)
			gitInfo.CommitDate = authorTimeFormatted
		case strings.HasPrefix(line, "filename "):
			gitInfo.File = strings.TrimPrefix(line, "filename ")
		}
	}

	if err := scanner.Err(); err != nil {
		// Ignore this error because it can only happen with the last line of the output
		// (the line git blame was called on) as the others are always short enough.
		// This line is not used by this function.
		if err.Error() != "bufio.Scanner: token too long" {
			return types.GitInfo{}, err
		}
	}

	return gitInfo, nil
}

func GetBranchName(directoryToScan string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = directoryToScan
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	latestCommitHash := strings.TrimSuffix(string(output), "\n")

	return latestCommitHash, nil
}

func GetLatestCommitHash(directoryToScan string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = directoryToScan
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	latestCommitHash := strings.TrimSuffix(string(output), "\n")

	return latestCommitHash, nil
}

func GetScannableObjectRemoteUrls(directoryToScan string) ([]string, error) {
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = directoryToScan
	gitRemoteOutput, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	scannableObjectRemoteUrls, err := parseGitRemoteOutput(gitRemoteOutput)
	if err != nil {
		return nil, err
	}

	return scannableObjectRemoteUrls, nil
}

func parseGitRemoteOutput(gitRemoteOutput []byte) ([]string, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(gitRemoteOutput)))
	scannableObjectRemoteUrls := make([]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		lineFields := strings.Fields(line)
		scannableObjectRemoteUrl := lineFields[1]

		if functional.ArrayIncludes(scannableObjectRemoteUrls, scannableObjectRemoteUrl) {
			continue
		}

		scannableObjectRemoteUrls = append(scannableObjectRemoteUrls, scannableObjectRemoteUrl)
	}

	if err := scanner.Err(); err != nil {
		// Ignore this error because it can only happen with the last line of the output
		// (the line git blame was called on) as the others are always short enough.
		// This line is not used by this function.
		if err.Error() != "bufio.Scanner: token too long" {
			return nil, err
		}
	}

	return scannableObjectRemoteUrls, nil
}
