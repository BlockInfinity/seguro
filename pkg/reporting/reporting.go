package reporting

import (
	"errors"
	"fmt"
	"net/http"

	resty "github.com/go-resty/resty/v2"
	"github.com/secguro/secguro-cli/pkg/config"
	"github.com/secguro/secguro-cli/pkg/types"
)

const endpointPostScan = "scans"

func ReportScan(authToken string, projectName string, revision string,
	unifiedFindings []types.UnifiedFinding) error {
	fmt.Print("Sending scan report to server...")

	urlEndpointPostScan := config.ServerUrl + "/" + endpointPostScan

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
