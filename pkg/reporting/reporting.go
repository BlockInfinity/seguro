package reporting

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	resty "github.com/go-resty/resty/v2"
	"github.com/secguro/secguro-cli/pkg/config"
	"github.com/secguro/secguro-cli/pkg/git"
	"github.com/secguro/secguro-cli/pkg/login"
	"github.com/secguro/secguro-cli/pkg/types"
)

const endpointPostScan = "scans"

func ReportScan(authToken string, scannableResourceName string, scannableResourceRemoteUrls []string,
	branch string, revision string, unifiedFindings []types.UnifiedFinding, failedDetectors []string) error {
	fmt.Print("Sending scan report to server...")

	authProvider := "secguro"

	urlEndpointPostScan := config.ServerUrl + "/" + endpointPostScan

	scanPostReq := types.ScanPostReq{
		ScannableResourceName:       scannableResourceName,
		ScannableResourceRemoteUrls: scannableResourceRemoteUrls,
		Branch:                      branch,
		Revision:                    revision,
		Findings:                    unifiedFindings,
		FailedDetectors:             failedDetectors,
	}

	result := types.ConfirmationRes{} //nolint: exhaustruct
	client := resty.New()
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", authProvider+" "+authToken).
		SetBody(scanPostReq).
		SetResult(&result).
		Post(urlEndpointPostScan)

	if err != nil {
		return err
	}

	if response.StatusCode() != http.StatusCreated {
		return errors.New("received bad status code")
	}

	if result.Status != "created" {
		return errors.New("received bad status response")
	}

	fmt.Println("done")

	return nil
}

func ReportScanIfApplicable(directoryToScan string,
	unifiedFindingsNotIgnored []types.UnifiedFinding, failedDetectors []string) error {
	authToken, err := login.GetAuthToken()
	if err != nil {
		return err
	}

	scannableResourceName, err := getScannableResourceName(directoryToScan)
	if err != nil {
		return err
	}

	branch, revision, scannableResourceRemoteUrls, err := getGitBasedScanMetadata(directoryToScan)
	if err != nil {
		return err
	}

	if authToken != "" {
		err = ReportScan(authToken, scannableResourceName, scannableResourceRemoteUrls,
			branch, revision, unifiedFindingsNotIgnored, failedDetectors)
		if err != nil {
			return err
		}
	}

	return nil
}

func getScannableResourceName(directoryToScan string) (string, error) {
	absPath, err := filepath.Abs(directoryToScan)
	if err != nil {
		return "", err
	}

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}

	var dirAbsPath string
	if fileInfo.IsDir() {
		dirAbsPath = absPath
	} else {
		dirAbsPath = filepath.Dir(absPath)
	}

	dirName := filepath.Base(dirAbsPath)

	return dirName, nil
}

func getGitBasedScanMetadata(directoryToScan string) (branch string,
	revision string, scannableResourceRemoteUrls []string, err error) {
	errorStringIfNotInGitRepo := "exit status 128"

	branch, err = git.GetBranchName(directoryToScan)
	if err != nil {
		// Set branch to empty string for paths that are not in git repos.
		if err.Error() == errorStringIfNotInGitRepo {
			branch = ""
		} else {
			return "", "", nil, err
		}
	}

	revision, err = git.GetLatestCommitHash(directoryToScan)
	if err != nil {
		// Set revision to empty string for paths that are not in git repos.
		if err.Error() == errorStringIfNotInGitRepo {
			revision = ""
		} else {
			return "", "", nil, err
		}
	}

	scannableResourceRemoteUrls, err = git.GetScannableResourceRemoteUrls(directoryToScan)
	if err != nil {
		// Set scannableResource remote URLs to empty array for paths that are not in git repos.
		if err.Error() == errorStringIfNotInGitRepo {
			scannableResourceRemoteUrls = make([]string, 0)
		} else {
			return "", "", nil, err
		}
	}

	err = nil

	return //nolint: nakedret
}
