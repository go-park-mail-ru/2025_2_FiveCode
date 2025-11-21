package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type Config struct {
	GRPCPort int `mapstructure:"grpc_port"`
	DB       DBConfig
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("notes.config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("config file not found, using environment variables only")
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cfg.DB.Host = v.GetString("DB_HOST")
	cfg.DB.Port = v.GetInt("DB_PORT")
	cfg.DB.User = v.GetString("DB_USER")
	cfg.DB.Password = v.GetString("DB_PASSWORD")
	cfg.DB.DBName = v.GetString("DB_NAME")
	cfg.DB.SSLMode = v.GetString("DB_SSLMODE")

	if cfg.GRPCPort == 0 {
		cfg.GRPCPort = v.GetInt("GRPC_PORT")
	}

	if cfg.DB.Host == "" {
		return nil, fmt.Errorf("DB_HOST is a required environment variable")
	}

	return &cfg, nil

}
