package dependencies

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

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

func downloadDependency(name string, fileNameExtension string, url string) error {
	const directoryPermissions = 0700
	err := os.MkdirAll(DependenciesDir, directoryPermissions)
	if err != nil {
		return err
	}

	return downloadFile(DependenciesDir+"/"+name+"."+fileNameExtension, url)
}

// https://stackoverflow.com/a/33853856
func downloadFile(filepath string, url string) error {
	// Create the file
	out, err := os.Create(filepath)
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
