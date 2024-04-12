package login

import (
	"errors"
	"os"
)

const deviceTokenFileName = "device_token"
const secguroConfigDirName = ".secguro"

func CommandLogin() error {
	// email address is auth token until auth method has been decided upon
	deviceToken := "bob@trashmail.com"

	pathSecguroConfigDir, err := getSecguroConfigDirPath()
	if err != nil {
		return err
	}

	err = ensureDirectoryExists(pathSecguroConfigDir)
	if err != nil {
		return err
	}

	const filePermissions = 0600
	err = os.WriteFile(pathSecguroConfigDir+"/"+deviceTokenFileName, []byte(deviceToken), filePermissions)
	if err != nil {
		return err
	}

	return nil
}

func GetDeviceToken() (string, error) {
	pathSecguroConfigDir, err := getSecguroConfigDirPath()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(pathSecguroConfigDir + "/" + deviceTokenFileName); err == nil {
		authTokenBytes, err := os.ReadFile(pathSecguroConfigDir + "/" + deviceTokenFileName)
		if err != nil {
			return "", err
		}

		authToken := string(authTokenBytes)

		return authToken, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return "", nil
	} else {
		return "", errors.New("cannot determine whether user is logged in")
	}
}

func ensureDirectoryExists(path string) error {
	const directoryPermissions = 0700
	return os.MkdirAll(path, directoryPermissions)
}

func getSecguroConfigDirPath() (string, error) {
	pathHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return pathHomeDir + "/" + secguroConfigDirName, nil
}
