package main

import (
	"errors"
	"fmt"
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
			return previousStep()
		}
	}

	return fixSecretStep2(func() error { return fixSecretStep1(previousStep, unifiedFinding) }, secret)
}

func fixSecretStep2(previousStep func() error, secret string) error {
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
		return fixSecretStep3(func() error { return fixSecretStep2(previousStep, secret) }, secret)
	case 2:
		return addSecretToIgnoreList(secret)
	}

	return errors.New("unexpected choice index")
}

func fixSecretStep3(previousStep func() error, secret string) error {
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
	if choiceIndex == -1 || choiceIndex == 0 {
		return previousStep()
	}

	if choiceIndex == 1 {
		return addSecretToIgnoreList(secret)
	}

	if choiceIndex == 2 {
		return fixSecretStepB3(func() error { return fixSecretStep3(previousStep, secret) }, secret)
	}

	if choiceIndex == 3 {
		return nil
	}

	return errors.New("unexpected choice index")
}

func addSecretToIgnoreList(secret string) error {
	fmt.Print("Adding secret to ignore list...")

	const filePermissions = 0644
	file, err := os.OpenFile(directoryToScan+"/"+secretsIgnoreFileName,
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

func fixSecretStepB3(previousStep func() error, secret string) error {
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
