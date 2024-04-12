package reporting

import (
	"errors"
	"fmt"
	"net/http"

	resty "github.com/go-resty/resty/v2"
	"github.com/secguro/secguro-cli/pkg/config"
	"github.com/secguro/secguro-cli/pkg/types"
)

const endpointSaveScan = "saveScan"

func ReportScan(unifiedFindings []types.UnifiedFinding) error {
	fmt.Print("Sending scan report to server...")

	urlEndpointSaveScan := config.ServerUrl + "/" + endpointSaveScan

	scanPostReq := ScanPostReq{
		ProjectName: "example project",
		Revision:    "example revision",
		Findings:    unifiedFindings,
	}

	result := ConfirmationRes{} //nolint: exhaustruct
	client := resty.New()
	// email address is auth token until auth method has been decided upon
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "bob@trashmail.com").
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
