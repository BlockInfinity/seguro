package dependencies

import "github.com/secguro/secguro-cli/pkg/utils"

func downloadAndExtractDependencycheck() error {
	filePath := DependenciesDir + "/" + "dependencycheck.zip"
	url := "https://github.com/jeremylong/DependencyCheck/releases/download/v9.0.9/dependency-check-9.0.9-release.zip"

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

	err = extractZipDependency("dependencycheck")

	return err
}
