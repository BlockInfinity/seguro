package utils

import (
	"errors"
	"os"
)

func DoesFileExist(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, errors.New("cannot determine whether file exists")
	}
}
