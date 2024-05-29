package dependencycheck

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/secguro/secguro-cli/pkg/config"
	"github.com/secguro/secguro-cli/pkg/types"
)

const endpointPostDependencycheckScan = "dependencycheckScans"

func getDependencycheckFindingsAsUnifiedFromServer(directoryToScan string,
	_gitMode bool) ([]types.UnifiedFinding, error) {
	manifestFiles, err := getManifestFiles(directoryToScan)
	if err != nil {
		return nil, err
	}

	urlEndpointPostDependencycheckScan := config.ServerUrl + "/" + endpointPostDependencycheckScan

	dependencycheckScanPostReq := types.DependencycheckScanPostReq{
		ManifestFiles: manifestFiles,
	}

	result := types.DependencycheckScanRes{} //nolint: exhaustruct
	client := resty.New()
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(dependencycheckScanPostReq).
		SetResult(&result).
		Post(urlEndpointPostDependencycheckScan)

	if err != nil {
		return nil, err
	}

	if response.StatusCode() != http.StatusOK {
		return nil, errors.New("received bad status code")
	}

	return result.UnifiedFindings, nil
}

func getManifestFiles(directoryToScan string) ([]types.FileReq, error) {
	absPathDirectoryToScan, err := filepath.Abs(directoryToScan)
	if err != nil {
		return nil, err
	}

	result := make([]types.FileReq, 0)

	err = filepath.WalkDir(absPathDirectoryToScan, func(path string, dirEntry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if dirEntry.IsDir() {
			return nil
		}

		if !strings.HasPrefix(path, absPathDirectoryToScan) {
			return errors.New("unexpected path (path path does not start with abs path of dir to scan)")
		}

		relativePath := strings.TrimPrefix(path, absPathDirectoryToScan)

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		filename := filepath.Base(relativePath)
		if !isManifestFile(filename) {
			return nil
		}

		// TODO: censor manifest files (e.g. remove the scripts attribute of a package.json file)
		// to only transfer dependencies
		fileReq := types.FileReq{
			Path:    relativePath,
			Content: content,
		}

		result = append(result, fileReq)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func isManifestFile(filename string) bool {
	switch filename {
	case "package.json", "package-lock.json":
		return true

	case "go.mod", "go.sum":
		return true

	default:
		return false
	}
}
