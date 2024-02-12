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

func extractGzDependency(name string) error {
	file, _ := os.Open(dependenciesDir + "/" + name + ".tar.gz")
	return extract.Gz(context.Background(), file, dependenciesDir+"/"+name, nil)
}

func downloadDependency(name string, url string) error {
	err := os.MkdirAll(dependenciesDir, os.ModePerm)
	if err != nil {
		return err
	}

	return downloadFile(dependenciesDir+"/"+name+".tar.gz", url)
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
