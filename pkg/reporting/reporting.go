package reporting

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	resty "github.com/go-resty/resty/v2"
	"github.com/secguro/secguro-cli/pkg/config"
	"github.com/secguro/secguro-cli/pkg/git"
	"github.com/secguro/secguro-cli/pkg/types"
)

const endpointSaveScan = "saveScan"

func ReportScan(authToken string, unifiedFindings []types.UnifiedFinding) error {
	fmt.Print("Sending scan report to server...")

	projectName := config.DirectoryToScan[strings.LastIndex(config.DirectoryToScan, "/")+1:]

	revision, err := git.GetLatestCommitHash()
	if err != nil {
		return err
	}

	urlEndpointSaveScan := config.ServerUrl + "/" + endpointSaveScan

	scanPostReq := ScanPostReq{
		ProjectName: projectName,
		Revision:    revision,
		Findings:    unifiedFindings,
	}

	result := ConfirmationRes{} //nolint: exhaustruct
	client := resty.New()
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", authToken).
		SetBody(scanPostReq).
		SetResult(&result).
		Post(urlEndpointSaveScan)

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
