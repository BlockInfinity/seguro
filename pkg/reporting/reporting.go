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

func ReportScan(authToken string, projectName string, projectRemoteUrls []string,
	revision string, unifiedFindings []types.UnifiedFinding, failedDetectors []string) error {
	fmt.Print("Sending scan report to server...")

	urlEndpointPostScan := config.ServerUrl + "/" + endpointPostScan

	scanPostReq := types.ScanPostReq{
		ProjectName:       projectName,
		ProjectRemoteUrls: projectRemoteUrls,
		Revision:          revision,
		Findings:          unifiedFindings,
		FailedDetectors:   failedDetectors,
	}

	result := types.ConfirmationRes{} //nolint: exhaustruct
	client := resty.New()
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", authToken).
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

	projectName, err := getProjectName(directoryToScan)
	if err != nil {
		return err
	}

	revision, err := git.GetLatestCommitHash(directoryToScan)
	if err != nil {
		// Set revision to empty string for paths that are not in git repos.
		if err.Error() == "exit status 128" {
			revision = ""
		} else {
			return err
		}
	}

	projectRemoteUrls, err := git.GetProjectRemoteUrls(directoryToScan)
	if err != nil {
		// Set project remote URLs to empty array for paths that are not in git repos.
		if err.Error() == "exit status 128" {
			projectRemoteUrls = make([]string, 0)
		} else {
			return err
		}
	}

	if authToken != "" {
		err = ReportScan(authToken, projectName, projectRemoteUrls,
			revision, unifiedFindingsNotIgnored, failedDetectors)
		if err != nil {
			return err
		}
	}

	return nil
}

func getProjectName(directoryToScan string) (string, error) {
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
