package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Meta_DependencycheckFinding struct {
	Dependencies []DependencycheckFinding
}

type DependencycheckFinding struct {
	FilePath        string
	Vulnerabilities []DependencycheckFinding_Vulnerabilities
}

type DependencycheckFinding_Vulnerabilities struct {
	Name string
}

func convertDependencycheckFindingToUnifiedFinding(dependencycheckFinding DependencycheckFinding,
	vulnerabilityIndex int) UnifiedFinding {
	const separator = "?"
	separatorIndex := strings.LastIndex(dependencycheckFinding.FilePath, separator)

	file := dependencycheckFinding.FilePath[:separatorIndex]
	packageAndVersionPossiblePrefixed := dependencycheckFinding.FilePath[separatorIndex+len(separator):]
	packageAndVersion := strings.TrimPrefix(packageAndVersionPossiblePrefixed, "/")

	return UnifiedFinding{
		Detector:           "dependencycheck",
		Rule:               dependencycheckFinding.Vulnerabilities[vulnerabilityIndex].Name,
		File:               file,
		LineStart:          -1,
		LineEnd:            -1,
		ColumnStart:        -1,
		ColumnEnd:          -1,
		Match:              packageAndVersion,
		Hint:               "",
		CommitHash:         "",
		CommitDate:         "",
		AuthorName:         "",
		AuthorEmailAddress: "",
		CommitSummary:      "",
	}
}

func getDependencycheckOutputJson(_gitMode bool) ([]byte, error) {
	dependencycheckOutputDirPath := dependenciesDir + "/dependencycheckOutput"
	dependencycheckOutputJsonPath := dependencycheckOutputDirPath + "/dependency-check-report.json"

	// secguro-ignore-next-line
	cmd := exec.Command(dependenciesDir+"/dependencycheck/dependency-check/bin/dependency-check.sh",
		"--scan", directoryToScan+"/**/package.json",
		"--scan", directoryToScan+"/**/package-lock.json",
		"--format", "JSON", "--out", dependencycheckOutputDirPath,
		"--nvdApiKey", os.Getenv("NVD_API_KEY"))
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Received output from dependencycheck:")
		fmt.Println(out)
		fmt.Println("Received error from dependencycheck:")
		fmt.Println(err)

		return nil, errors.New("dependencycheck failed")
	}

	if out == nil {
		return nil, errors.New("did not receive output from dependencycheck")
	}

	dependencycheckOutputJson, err := os.ReadFile(dependencycheckOutputJsonPath)

	return dependencycheckOutputJson, err
}

func getDependencycheckFindingsAsUnified(gitMode bool) ([]UnifiedFinding, error) {
	dependencycheckOutputJson, err := getDependencycheckOutputJson(gitMode)
	if err != nil {
		return nil, err
	}

	var metaDependencycheckFindings Meta_DependencycheckFinding
	err = json.Unmarshal(dependencycheckOutputJson, &metaDependencycheckFindings)
	if err != nil {
		return nil, err
	}

	dependencycheckFindings := metaDependencycheckFindings.Dependencies
	unifiedFindings := make([]UnifiedFinding, 0)
	for _, dependencycheckFinding := range dependencycheckFindings {
		for vulnerabilityIndex := range dependencycheckFinding.Vulnerabilities {
			unifiedFinding := convertDependencycheckFindingToUnifiedFinding(
				dependencycheckFinding, vulnerabilityIndex)
			unifiedFindings = append(unifiedFindings, unifiedFinding)
		}
	}

	return unifiedFindings, nil
}

func downloadAndExtractDependencycheck() error {
	err := downloadDependency("dependencycheck", "zip",
		"https://github.com/jeremylong/DependencyCheck/releases/download/v9.0.9/dependency-check-9.0.9-release.zip")
	if err != nil {
		return err
	}

	err = extractZipDependency("dependencycheck")

	return err
}
