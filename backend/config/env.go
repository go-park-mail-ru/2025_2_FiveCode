package config

import (
	"errors"
	"github.com/joho/godotenv"

	"github.com/rs/zerolog/log"
	"os"
)

func ReadServerAddress() (string, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Info().Msgf("Can not load .env file %v", err)
	}

	serverHost := os.Getenv("SERVER_HOST")
	serverPort := os.Getenv("SERVER_PORT")

	if serverHost == "" || serverPort == "" {
		return "", errors.New("SERVER_HOST and SERVER_PORT environment variables not set")
	}

	return serverHost + ":" + serverPort, nil
}

func ReadConfigPath() (string, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Info().Msgf("Can not load .env file %v", err)
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		return "", errors.New("CONFIG_PATH environment variable not set")
	}

	return configPath, nil
}
