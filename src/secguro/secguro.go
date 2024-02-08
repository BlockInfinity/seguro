package main

import (
	"fmt"
	"os"
	"os/exec"
)

// const directoryToScan = "/home/christoph/Development/Work/wallet/"
const directoryToScan = "."

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
	if out == nil {
		panic("did not receive output from gitleaks")
	}

	fmt.Println("Output by gitleaks:")
	fmt.Println(string(out[:]))

	os.Exit(0)
}
