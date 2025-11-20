package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"github.com/rs/zerolog/log"
)

func loadEnvFile() error {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AddConfigPath(".")
	viper.AddConfigPath("../")
	viper.AddConfigPath("../../")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info().Msg("'.env' file not found, using environment variables.")
		} else {
			log.Error().Err(err).Msg("error reading env file")
			return err
		}
	}

	return nil
}

func ReadServerAddress() (string, error) {
	err := loadEnvFile()
	if err != nil {
		return "", fmt.Errorf("failed to load env file: %w", err)
	}

	serverHost := viper.GetString("SERVER_HOST")
	serverPort := viper.GetString("SERVER_PORT")

	if serverHost == "" || serverPort == "" {
		return "", errors.New("SERVER_HOST or SERVER_PORT environment variables not set")
	}

	return serverHost + ":" + serverPort, nil
}

func ReadConfigPath() (string, error) {
	err := loadEnvFile()
	if err != nil {
		return "", fmt.Errorf("failed to load env file: %w", err)
	}

	configPath := viper.GetString("CONFIG_PATH")
	if configPath == "" {
		return "", errors.New("CONFIG_PATH environment variable not set")
	}

	return configPath, nil
}
