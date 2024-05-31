package dependencies

import (
	"os"

	"github.com/secguro/secguro-cli/pkg/config"
	"github.com/secguro/secguro-cli/pkg/functional"
)

func InstallDependencies(disabledDetectors []string) error {
	if !functional.ArrayIncludes(disabledDetectors, "gitleaks") {
		err := downloadAndExtractGitleaks()
		if err != nil {
			return err
		}
	}

	if !functional.ArrayIncludes(disabledDetectors, "semgrep") {
		err := installSemgrep()
		if err != nil {
			return err
		}
	}

	usingDependencycheckOnServer := os.Getenv(config.NvdApiKeyEnvVarName) == ""
	if !functional.ArrayIncludes(disabledDetectors, "dependencycheck") && !usingDependencycheckOnServer {
		err := downloadAndExtractDependencycheck()
		if err != nil {
			return err
		}
	}

	return nil
}
