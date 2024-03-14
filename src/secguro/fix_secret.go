package main

import (
	"os"
	"os/exec"
)

func fixSecret(previousStep func() error, unifiedFinding UnifiedFinding) error {
	return fixSecretStep1(previousStep, unifiedFinding)
}

func fixSecretStep1(previousStep func() error, unifiedFinding UnifiedFinding) error {
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
			previousStep()
		}
	}

	return fixSecretStep2(func() error { return fixSecretStep1(previousStep, unifiedFinding) }, secret)
}

func fixSecretStep2(previousStep func() error, secret string) error {
	for {
		searchResult, err := findStringInGitIndex(secret)
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
			"You may use environmnt variables to insert secrets into your " +
			"program. Make sure that your secrets are also available in any " +
			"CI tools you are using, as well as your production and CD environments." +
			"Check out: https://docs.github.com/en/actions/security-guides/using-secrets-in-github-actions" +
			"\n\n" +
			"We found the secret in:\n" +
			searchResult
		choices := []string{"back", "I have removed the secret from the latest commit."}
		choiceIndex, _, err := getOptionChoice(prompt, choices)
		if err != nil {
			return err
		}
		if choiceIndex == -1 || choiceIndex == 0 {
			return previousStep()
		}
	}

	return removeSecret(secret)
}

func findStringInGitIndex(secret string) (string, error) {
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

func downloadBfg() error {
	err := downloadDependency("bfg", "jar",
		"https://repo1.maven.org/maven2/com/madgag/bfg/1.14.0/bfg-1.14.0.jar")
	return err
}

func removeSecret(secret string) error {
	err := downloadBfg()
	if err != nil {
		return err
	}

	pathReplacementsFile := dependenciesDir + "/" + "replacements"

	const filePermissions = 0600
	defer os.Remove(pathReplacementsFile)
	// TODO: find out how escaping works with bfg
	// (seems not to be document; single and double quotes do not work)
	err = os.WriteFile(pathReplacementsFile, []byte(secret), filePermissions)
	if err != nil {
		return err
	}

	pathBfg := dependenciesDir + "/bfg.jar"

	cmd := exec.Command("java", "-jar", pathBfg, "--replace-text", pathReplacementsFile, ".")
	cmd.Dir = directoryToScan
	_, err = cmd.Output()

	if err != nil {
		return err
	}

	return nil
}
