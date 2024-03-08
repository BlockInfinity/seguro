package main

import (
	"fmt"
	"os/exec"
)

func fixSecret(unifiedFinding UnifiedFinding) error {
	fmt.Println("would fix:")
	fmt.Println(unifiedFinding)

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

	fmt.Println("got:")
	fmt.Println(secret)

	return nil
}
