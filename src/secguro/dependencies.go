package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/codeclysm/extract/v3"
)

const dependenciesDir = "/tmp/secguroDependencies"

func extractDependencies() error {
	file, _ := os.Open(dependenciesDir + "/gitleaks.tar.gz")
	return extract.Gz(context.Background(), file, dependenciesDir+"/gitleaks", nil)
}

func downloadDependencies() error {
	err := os.MkdirAll(dependenciesDir, os.ModePerm)
	if err != nil {
		return err
	}

	return downloadFile(dependenciesDir+"/gitleaks.tar.gz", "https://github.com/gitleaks/gitleaks/releases/download/v8.18.2/gitleaks_8.18.2_linux_x64.tar.gz")
}

// https://stackoverflow.com/a/33853856
func downloadFile(filepath string, url string) (err error) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
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
