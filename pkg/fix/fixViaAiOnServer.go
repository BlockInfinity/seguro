package fix

import (
	"errors"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/secguro/secguro-cli/pkg/config"
	"github.com/secguro/secguro-cli/pkg/types"
)

const endpointPostFixedFileContent = "fixedFileContents"

func getFixedFileContentFromChatGptFromServer(fileContent string,
	problemLineNumber int, hint string) (string, error) {
	urlEndpointPostFixedFileContent := config.ServerUrl + "/" + endpointPostFixedFileContent

	fixedFileContentPostReq := types.FixedFileContentPostReq{
		FileContent:       fileContent,
		ProblemLineNumber: problemLineNumber,
		Hint:              hint,
	}

	result := types.FixedFileContentRes{} //nolint: exhaustruct
	client := resty.New()
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(fixedFileContentPostReq).
		SetResult(&result).
		Post(urlEndpointPostFixedFileContent)

	if err != nil {
		return "", err
	}

	if response.StatusCode() != http.StatusOK {
		return "", errors.New("received bad status code")
	}

	return result.FixedFileContent, nil
}
