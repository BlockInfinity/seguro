package dependencies

import (
	"errors"
	"runtime"

	"github.com/secguro/secguro-cli/pkg/utils"
)

func downloadAndExtractGitleaks() error {
	var url string
	switch runtime.GOOS {
	case "linux":
		url = "https://github.com/gitleaks/gitleaks/releases/download/v8.18.3/gitleaks_8.18.3_linux_x64.tar.gz"
	case "darwin":
		url = "https://github.com/gitleaks/gitleaks/releases/download/v8.18.3/gitleaks_8.18.3_darwin_arm64.tar.gz"
	default:
		return errors.New("Unsupported platform")
	}

	filePath := DependenciesDir + "/" + "gitleaks.tar.gz"

	doesAlreadExist, err := utils.DoesFileExist(filePath)
	if err != nil {
		return err
	}
	if doesAlreadExist {
		return nil
	}

	err = downloadDependency(filePath, url)
	if err != nil {
		return err
	}

	err = extractGzDependency("gitleaks")

	return err
}
