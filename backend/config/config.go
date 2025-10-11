package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type CorsConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type Config struct {
	Cors CorsConfig `yaml:"cors"`
}

func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error reading config file, %s", err)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("unable to decode config, %v", err)
	}

	return &config, nil
}
