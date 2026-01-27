package config

import (
	"fmt"
	"os"
)

func GetEnvVariable(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("environment variable %s not set", key)
	}

	return value, nil
}

func GetEnvVariableOrDefault(key string, def string) string {
	value := os.Getenv(key)
	if value == "" {
		return def
	}

	return value
}
