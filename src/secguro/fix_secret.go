package main

import (
	"fmt"
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

	secretIsInIndex := true
	for secretIsInIndex {
		searchResult, err := findStringInGitIndex(secret)
		if err != nil {
			return err
		}
		secretIsInIndex = len(searchResult) > 0

		if secretIsInIndex {
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
	}

	fmt.Println("would fix:")
	fmt.Println(unifiedFinding)

	return nil
}

func findStringInGitIndex(secret string) (string, error) {
	cmd := exec.Command("git", "grep", "--color", secret, "HEAD", "--", ".")
	cmd.Dir = directoryToScan
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
