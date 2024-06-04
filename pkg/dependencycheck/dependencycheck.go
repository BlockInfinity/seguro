package dependencycheck

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/secguro/secguro-cli/pkg/config"
	"github.com/secguro/secguro-cli/pkg/dependencies"
	"github.com/secguro/secguro-cli/pkg/types"
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

func convertDependencycheckFindingToUnifiedFinding(directoryToScan string,
	dependencycheckFinding DependencycheckFinding, vulnerabilityIndex int) types.UnifiedFinding {
	// dependencycheck uses "?" for npm dependencies but ":" or none at all for go dependencies.
	separator := "?"
	separatorIndex := strings.LastIndex(dependencycheckFinding.FilePath, separator)
	if separatorIndex == -1 {
		separator = ":"
		separatorIndex = strings.LastIndex(dependencycheckFinding.FilePath, separator)
	}

	// Prevent crashing in the case where there is no separator. Only seems to happen
	// in experimental mode. Problem not handled better because it might go away.
	if separatorIndex == -1 {
		separatorIndex = len(dependencycheckFinding.FilePath) - 1
	}

	fileFullPath := dependencycheckFinding.FilePath[:separatorIndex]
	// Contrary to the other detectors, dependencycheck returns the path
	// including the path of the directory to scan.
	file := strings.TrimPrefix(fileFullPath, directoryToScan)
	packageAndVersionPossiblePrefixed := dependencycheckFinding.FilePath[separatorIndex+len(separator):]
	packageAndVersion := strings.TrimPrefix(packageAndVersionPossiblePrefixed, "/")

	return types.UnifiedFinding{
		Detector:    "dependencycheck",
		Rule:        dependencycheckFinding.Vulnerabilities[vulnerabilityIndex].Name,
		File:        file,
		LineStart:   -1,
		LineEnd:     -1,
		ColumnStart: -1,
		ColumnEnd:   -1,
		Match:       packageAndVersion,
		Hint:        "",
		Severity:    "WARNING", // TODO: differentiate severity for dependencycheck
		GitInfo:     nil,
	}
}

func getDependencycheckOutputJson(directoryToScan string, _gitMode bool) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)
	dependencycheckOutputDirPath := tmpDir
	dependencycheckOutputJsonPath := dependencycheckOutputDirPath + "/dependency-check-report.json"

	// secguro-ignore-next-line
	cmd := exec.Command(dependencies.DependenciesDir+"/dependencycheck/dependency-check/bin/dependency-check.sh",
		"--enableExperimental", // necessary for support of go dependencies
		"--scan", directoryToScan+"/**/package.json",
		"--scan", directoryToScan+"/**/package-lock.json",
		"--scan", directoryToScan+"/**/go.mod", // .sum files are not considered by dependencycheck
		"--format", "JSON", "--out", dependencycheckOutputDirPath,
		"--nvdApiKey", os.Getenv(config.NvdApiKeyEnvVarName))
	out, err := cmd.Output()
	if err != nil {
		if !config.TolerateDependecycheckErrorExitCodes {
			fmt.Println("Received output from dependencycheck:")
			fmt.Println(out)
			fmt.Println("Received error from dependencycheck:")
			fmt.Println(err)

			return nil, errors.New("dependencycheck failed")
		}

		fmt.Println("Received error from dependencycheck but continuing anyway...")
	}

	if out == nil {
		return nil, errors.New("did not receive output from dependencycheck")
	}

	dependencycheckOutputJson, err := os.ReadFile(dependencycheckOutputJsonPath)

	return dependencycheckOutputJson, err
}

func getDependencycheckFindingsAsUnifiedLocally(directoryToScan string,
	gitMode bool) ([]types.UnifiedFinding, error) {
	dependencycheckOutputJson, err := getDependencycheckOutputJson(directoryToScan, gitMode)
	if err != nil {
		return nil, err
	}

	var metaDependencycheckFindings Meta_DependencycheckFinding
	err = json.Unmarshal(dependencycheckOutputJson, &metaDependencycheckFindings)
	if err != nil {
		return nil, err
	}

	dependencycheckFindings := metaDependencycheckFindings.Dependencies
	unifiedFindings := make([]types.UnifiedFinding, 0)
	for _, dependencycheckFinding := range dependencycheckFindings {
		for vulnerabilityIndex := range dependencycheckFinding.Vulnerabilities {
			unifiedFinding := convertDependencycheckFindingToUnifiedFinding(directoryToScan,
				dependencycheckFinding, vulnerabilityIndex)
			unifiedFindings = append(unifiedFindings, unifiedFinding)
		}
	}

	return unifiedFindings, nil
}

func GetDependencycheckFindingsAsUnified(directoryToScan string, gitMode bool,
	unifiedFindingsChannel chan types.UnifiedFinding,
	detectorTerminationChannel chan types.DetectorTermination) {
	var f func(directoryToScan string, gitMode bool) ([]types.UnifiedFinding, error)
	if os.Getenv(config.NvdApiKeyEnvVarName) == "" {
		f = getDependencycheckFindingsAsUnifiedFromServer
	} else {
		f = getDependencycheckFindingsAsUnifiedLocally
	}

	unifiedFindings, err := f(directoryToScan, gitMode)
	if err != nil {
		fmt.Println(err)
		detectorTerminationChannel <- types.DetectorTermination{
			Detector:   "dependencycheck",
			Successful: false,
		}
	}

	for _, unifiedFinding := range unifiedFindings {
		unifiedFindingsChannel <- unifiedFinding
	}

	detectorTerminationChannel <- types.DetectorTermination{
		Detector:   "dependencycheck",
		Successful: true,
	}
}
