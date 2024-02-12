package main

import (
	"errors"
	"os/exec"
)

func installSemgrep() error {
	cmd := exec.Command("python3", "-m", "pipx", "install", "semgrep")
	_, err := cmd.Output()
	if err != nil {
		return errors.New("Failed to install Semgrep. Make sure that python3 and pipx are installed.")
	}

	return nil
}
