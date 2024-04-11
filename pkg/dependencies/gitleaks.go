package dependencies

import (
	"errors"
	"runtime"
)

func downloadAndExtractGitleaks() error {
	var url string
	switch runtime.GOOS {
	case "linux":
		url = "https://github.com/gitleaks/gitleaks/releases/download/v8.18.2/gitleaks_8.18.2_linux_x64.tar.gz"
	case "darwin":
		url = "https://github.com/gitleaks/gitleaks/releases/download/v8.18.2/gitleaks_8.18.2_darwin_arm64.tar.gz"
	default:
		return errors.New("Unsupported platform")
	}

	err := downloadDependency("gitleaks", "tar.gz", url)
	if err != nil {
		return err
	}

	err = extractGzDependency("gitleaks")

	return err
}
