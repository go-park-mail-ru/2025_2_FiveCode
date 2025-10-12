package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type CorsConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type CookieConfig struct {
	SessionDuration int `mapstructure:"session_duration"`
}

type Config struct {
	Cors   CorsConfig   `mapstructure:"cors"`
	Cookie CookieConfig `mapstructure:"cookie"`
}

func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "read config failed")
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal config failed")
	}

	return &config, nil
}
