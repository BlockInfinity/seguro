package main

import (
	"fmt"
	"os"
	"os/exec"
)

// TODO: turn into CLI parameter
const directoryToScan = "/home/christoph/Development/Work/wallet/"

func main() {
	err := downloadDependencies()
	if err != nil {
		panic(err)
	}

	err = extractDependencies()
	if err != nil {
		panic(err)
	}

	cmd := exec.Command(dependenciesDir+"/gitleaks/gitleaks", "detect", "-v")
	cmd.Dir = directoryToScan
	// Ignore error because this is expected to deliver an exict code not equal to 0.
	out, _ := cmd.Output()

	fmt.Println(string(out[:]))

	os.Exit(0)
}
