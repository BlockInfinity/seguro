package login

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/secguro/secguro-cli/pkg/config"
	"github.com/secguro/secguro-cli/pkg/types"
)

const ciTokenEnvVarName = "SECGURO_CI_TOKEN"

const deviceTokenFileName = "device_token"
const secguroConfigDirName = ".secguro"

const endpointPostDevice = "devices"
const endpointGetDeviceRegistration = "devices/{deviceId}/registrations"

func CommandLogin() error {
	deviceId, deviceToken, err := acquireDeviceIdAndDeviceToken()
	if err != nil {
		return err
	}

	loginUrl := fmt.Sprintf("http://localhost:3000/devices?deviceRegistrationId=%d", deviceId)
	fmt.Println("Please follow this link to register this device: " + loginUrl)

	for {
		time.Sleep(config.DeviceRegistrationPollingFrequencyInMs * time.Millisecond)

		isRegistered, err := isDeviceRegistered(deviceId)
		if err != nil {
			return err
		}

		if isRegistered {
			break
		}
	}

	err = saveDeviceToken(deviceToken)
	if err != nil {
		return err
	}

	return nil
}

func saveDeviceToken(deviceToken string) error {
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

func acquireDeviceIdAndDeviceToken() (uint, string, error) {
	deviceName, err := os.Hostname()
	if err != nil {
		return 0, "", err
	}

	urlEndpointPostDevice := config.ServerUrl + "/" + endpointPostDevice

	devicePostReq := types.DevicePostReq{
		DeviceName: deviceName,
	}

	result := types.DevicePostRes{} //nolint: exhaustruct
	client := resty.New()
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(devicePostReq).
		SetResult(&result).
		Post(urlEndpointPostDevice)

	if err != nil {
		return 0, "", err
	}

	if response.StatusCode() != http.StatusCreated {
		return 0, "", errors.New("received bad status code")
	}

	return result.ID, result.DeviceToken, nil
}

func isDeviceRegistered(deviceId uint) (bool, error) {
	urlEndpointGetDeviceRegistration := strings.Replace(
		config.ServerUrl+"/"+endpointGetDeviceRegistration,
		"{deviceId}", fmt.Sprintf("%d", deviceId), 1)

	result := types.DeviceRegistrationRes{} //nolint: exhaustruct
	client := resty.New()
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&result).
		Get(urlEndpointGetDeviceRegistration)

	if err != nil {
		return false, err
	}

	if response.StatusCode() != http.StatusOK {
		return false, errors.New("received bad status code")
	}

	return result.IsRegistered, nil
}

func GetAuthToken() (string, error) {
	ciToken := os.Getenv(ciTokenEnvVarName)
	if ciToken != "" {
		return ciToken, nil
	}

	return getDeviceToken()
}

func getDeviceToken() (string, error) {
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
