package dependencies

func downloadAndExtractDependencycheck() error {
	filePath := DependenciesDir + "/" + "dependencycheck.zip"

	err := downloadDependency(filePath,
		"https://github.com/jeremylong/DependencyCheck/releases/download/v9.0.9/dependency-check-9.0.9-release.zip")
	if err != nil {
		return err
	}

	err = extractZipDependency("dependencycheck")

	return err
}
