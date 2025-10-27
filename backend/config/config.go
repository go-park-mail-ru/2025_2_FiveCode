package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type CorsConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type CookieConfig struct {
	SessionDuration int `mapstructure:"session_duration"`
}

type MinIOConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	BucketName      string `mapstructure:"bucket_name"`
}

type Config struct {
	Cors   CorsConfig   `mapstructure:"cors"`
	Cookie CookieConfig `mapstructure:"cookie"`
	MinIO  MinIOConfig  `mapstructure:"minio"`
}

func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
