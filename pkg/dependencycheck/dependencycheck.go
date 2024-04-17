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

const NvdApiKeyEnvVarName = "NVD_API_KEY"

func convertDependencycheckFindingToUnifiedFinding(dependencycheckFinding DependencycheckFinding,
	vulnerabilityIndex int) types.UnifiedFinding {
	// dependencycheck uses "?" for npm dependencies but ":" or none at att for go dependencies.
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

	file := dependencycheckFinding.FilePath[:separatorIndex]
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
		GitInfo:     nil,
	}
}

func getDependencycheckOutputJson(directoryToScan string, _gitMode bool) ([]byte, error) {
	dependencycheckOutputDirPath := dependencies.DependenciesDir + "/dependencycheckOutput"
	dependencycheckOutputJsonPath := dependencycheckOutputDirPath + "/dependency-check-report.json"

	// secguro-ignore-next-line
	cmd := exec.Command(dependencies.DependenciesDir+"/dependencycheck/dependency-check/bin/dependency-check.sh",
		"--enableExperimental", // necessary for support of go dependencies
		"--scan", directoryToScan+"/**/package.json",
		"--scan", directoryToScan+"/**/package-lock.json",
		"--scan", directoryToScan+"/**/go.mod", // .sum files are not considered by dependencycheck
		"--format", "JSON", "--out", dependencycheckOutputDirPath,
		"--nvdApiKey", os.Getenv(NvdApiKeyEnvVarName))
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

func GetDependencycheckFindingsAsUnified(directoryToScan string, gitMode bool) ([]types.UnifiedFinding, error) {
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
			unifiedFinding := convertDependencycheckFindingToUnifiedFinding(
				dependencycheckFinding, vulnerabilityIndex)
			unifiedFindings = append(unifiedFindings, unifiedFinding)
		}
	}

	return unifiedFindings, nil
}
