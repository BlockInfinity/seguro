package dependencies

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/codeclysm/extract/v3"
)

const DependenciesDir = "/tmp/secguroDependencies"

func extractGzDependency(name string) error {
	file, _ := os.Open(DependenciesDir + "/" + name + ".tar.gz")
	return extract.Gz(context.Background(), file, DependenciesDir+"/"+name, nil)
}

func extractZipDependency(name string) error {
	file, _ := os.Open(DependenciesDir + "/" + name + ".zip")
	return extract.Zip(context.Background(), file, DependenciesDir+"/"+name, nil)
}

func downloadDependency(filePath string, url string) error {
	dirPath := filepath.Dir(filePath)

	const directoryPermissions = 0700
	err := os.MkdirAll(dirPath, directoryPermissions)
	if err != nil {
		return err
	}

	return downloadFile(filePath, url)
}

// https://stackoverflow.com/a/33853856
func downloadFile(filePath string, url string) error {
	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url) //nolint: noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
