package dependencies

import "github.com/secguro/secguro-cli/pkg/utils"

func DownloadBfg() error {
	filePath := DependenciesDir + "/" + "bfg.jar"
	url := "https://repo1.maven.org/maven2/com/madgag/bfg/1.14.0/bfg-1.14.0.jar"

	doesAlreadExist, err := utils.DoesFileExist(filePath)
	if err != nil {
		return err
	}
	if doesAlreadExist {
		return nil
	}

	err = downloadDependency(filePath, url)

	return err
}
