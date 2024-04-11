package dependencies

import (
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

	if !functional.ArrayIncludes(disabledDetectors, "dependencycheck") {
		err := downloadAndExtractDependencycheck()
		if err != nil {
			return err
		}
	}

	return nil
}
