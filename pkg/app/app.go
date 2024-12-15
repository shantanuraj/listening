package dirs

import (
	"os"
	"path"
)

const APP_NAME = "listening"

func CacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cacheDir := path.Join(homeDir, ".cache", APP_NAME)
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		os.MkdirAll(cacheDir, 0755)
	}

	return cacheDir, nil
}

func CredentialsPath() (string, error) {
	cacheDir, err := CacheDir()
	if err != nil {
		return "", err
	}

	return path.Join(cacheDir, "credentials.json"), nil
}
