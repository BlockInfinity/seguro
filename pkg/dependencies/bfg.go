package dependencies

func DownloadBfg() error {
	filePath := DependenciesDir + "/" + "bfg.jar"

	err := downloadDependency(filePath,
		"https://repo1.maven.org/maven2/com/madgag/bfg/1.14.0/bfg-1.14.0.jar")

	return err
}
