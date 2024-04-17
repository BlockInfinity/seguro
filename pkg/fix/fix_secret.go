package fix

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/secguro/secguro-cli/pkg/dependencies"
	"github.com/secguro/secguro-cli/pkg/ignoring"
	"github.com/secguro/secguro-cli/pkg/types"
)

func fixSecret(directoryToScan string, previousStep func() error, unifiedFinding types.UnifiedFinding) error {
	return fixSecretStep1(directoryToScan, previousStep, unifiedFinding)
}

func fixSecretStep1(directoryToScan string,
	previousStep func() error, unifiedFinding types.UnifiedFinding) error {
	prompt := "Please specify the secret in question. " +
		"Note that we are not always able to determine the exact bounds of " +
		"the secret, so it's important you specify the secret exactly."

	secret := ""
	for secret == "" {
		var goBack bool
		var err error
		secret, goBack, err = getTextInput(prompt, unifiedFinding.Match)
		if err != nil {
			return err
		}
		if goBack {
			return previousStep()
		}
	}

	return fixSecretStep2(directoryToScan,
		func() error { return fixSecretStep1(directoryToScan, previousStep, unifiedFinding) }, secret)
}

func fixSecretStep2(directoryToScan string, previousStep func() error, secret string) error {
	prompt := "Secret: " + secret +
		"\n\n" +
		"Replace how you access this secret now." +
		"\n\n" +
		"You may use environmnt variables to insert secrets into your " +
		"program. Make sure that your secrets are also available in any " +
		"CI tools you are using, as well as your production and CD environments. " +
		"Check out: https://docs.github.com/en/actions/security-guides/using-secrets-in-github-actions" +
		"\n\n" +
		"If you can change or invalidate this secret, we recommend that you do so."

	choices := []string{"back", "continue", "This is a false positive. Add it to the ignore list."}
	choiceIndex, err := getOptionChoice(prompt, choices)
	if err != nil {
		return err
	}

	switch choiceIndex {
	case -1, 0:
		return previousStep()
	case 1:
		return fixSecretStep3(directoryToScan,
			func() error { return fixSecretStep2(directoryToScan, previousStep, secret) }, secret)
	case 2:
		return addSecretToIgnoreList(directoryToScan, secret)
	}

	return errors.New("unexpected choice index")
}

func fixSecretStep3(directoryToScan string, previousStep func() error, secret string) error {
	prompt := "Secret: " + secret +
		"\n\n" +
		"If you were not able to invalidate the secret, the git history needs " +
		"to be re-written to remove the secret from it. In this case, please " +
		"take the necessary precautions: \n" +
		"• make sure that the latest revision does not contain the secret anymore\n" +
		"• make sure that your production deployment as well as your CI can access the secret through other means\n" +
		"• merge all pull requests\n" +
		"• merge all local branches\n" +
		"• make sure that your team members merge all of their local branches\n" +
		"• make sure that you and your team members all pull and are on the same revision" +
		"\n\n" +
		"After the secret has been removed from the git history, you will need to force-push " +
		"and your team members will need to force-pull."

	choices := []string{
		"back",
		"This secret is not valid anymore. Add it to the ignore list.",
		"Remove the secret from the git history. (Avoids future detection too.)",
		"Do nothing.",
	}
	choiceIndex, err := getOptionChoice(prompt, choices)
	if err != nil {
		return err
	}
	switch choiceIndex {
	case -1, 0:
		return previousStep()
	case 1:
		return addSecretToIgnoreList(directoryToScan, secret)
	case 2:
		return fixSecretStepB3(directoryToScan,
			func() error { return fixSecretStep3(directoryToScan, previousStep, secret) }, secret)
	case 3:
		return nil
	}

	return errors.New("unexpected choice index")
}

func addSecretToIgnoreList(directoryToScan string, secret string) error {
	fmt.Print("Adding secret to ignore list...")

	const filePermissions = 0644
	file, err := os.OpenFile(directoryToScan+"/"+ignoring.SecretsIgnoreFileName,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermissions)
	if err != nil {
		return err
	}
	if _, err := file.WriteString("\n" + secret); err != nil {
		file.Close() // ignore error; Write error takes precedence
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	fmt.Println("done")

	return nil
}

func fixSecretStepB3(directoryToScan string, previousStep func() error, secret string) error {
	for {
		searchResult, err := findStringInGitIndex(directoryToScan, secret)
		if err != nil {
			return err
		}
		secretIsInIndex := len(searchResult) > 0

		if !secretIsInIndex {
			break
		}

		prompt := "The specified secret is in the git index. Please replace the " +
			"secret, commit your changes, and try again. We only delete " +
			"secrets that are not in the git index to make sure that your " +
			"code keeps working. The file system state of your latest commit " +
			"will never be modified by secguro when deleting secrets." +
			"\n\n" +

			"We found the secret in:\n" +
			searchResult
		choices := []string{"back", "I have removed the secret from the latest commit."}
		choiceIndex, err := getOptionChoice(prompt, choices)
		if err != nil {
			return err
		}
		if choiceIndex == -1 || choiceIndex == 0 {
			return previousStep()
		}
	}

	return removeSecret(directoryToScan, secret)
}

func findStringInGitIndex(directoryToScan string, secret string) (string, error) {
	cmd := exec.Command("git", "grep", "--color", secret, "HEAD", "--", ".")
	cmd.Dir = directoryToScan
	out, err := cmd.Output()
	if err != nil {
		// This is expected to happen when there are no search results.
		if err.Error() == "exit status 1" {
			return string(out), nil
		}

		return "", err
	}

	return string(out), nil
}

func removeSecret(directoryToScan string, secret string) error {
	err := dependencies.DownloadBfg()
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	pathReplacementsFile := tmpDir + "/" + "replacements"

	const filePermissions = 0600
	defer os.Remove(pathReplacementsFile)
	// TODO: find out how escaping works with bfg
	// (seems not to be document; single and double quotes do not work)
	err = os.WriteFile(pathReplacementsFile, []byte(secret), filePermissions)
	if err != nil {
		return err
	}

	pathBfg := dependencies.DependenciesDir + "/bfg.jar"

	cmd := exec.Command("java", "-jar", pathBfg, "--replace-text", pathReplacementsFile, ".")
	cmd.Dir = directoryToScan
	_, err = cmd.Output()

	if err != nil {
		return err
	}

	return nil
}
