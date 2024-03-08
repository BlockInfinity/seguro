package main

import (
	"os"
	"os/exec"
)

func fixSecret(unifiedFinding UnifiedFinding) error {
	// TODO: make line breaks dependent on the terminal's width
	prompt := "Please specify the secret to remove from all git history. \n" +
		"Note that we are not always able to determine the exact bounds of \n" +
		"the secret, so it's important you specify the secret exactly."

	secret := ""
	for secret == "" {
		var err error
		secret, err = getTextInput(prompt, unifiedFinding.Match)
		if err != nil {
			return err
		}
	}

	for {
		searchResult, err := findStringInGitIndex(secret)
		if err != nil {
			return err
		}
		secretIsInIndex := len(searchResult) > 0

		if !secretIsInIndex {
			break
		}

		prompt = "The specified secret is in the git index. Please replace the \n" +
			"secret, commit your changes, and try again. We only delete \n" +
			"secrets that are not in the git index to make sure that your \n" +
			"code keeps working. The file system state of your latest commit \n" +
			"will never be modified by secguro when deleting secrets.\n" +
			"\n" +
			"You may use environmnt variables to insert secrets into your \n" +
			"program. Make sure that your secrets are also available in any \n" +
			"CI tools you are using, as well as your production and CD environments.\n" +
			"Check out: https://docs.github.com/en/actions/security-guides/using-secrets-in-github-actions\n" +
			"\n" +
			"We found the secret in:\n" +
			searchResult
		choices := []string{"back", "I have removed the secret from the latest commit."}
		choiceIndex, _, err := getOptionChoice(prompt, choices)
		if err != nil {
			return err
		}
		if choiceIndex == 0 {
			return nil
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
