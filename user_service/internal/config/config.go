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
	GRPCPort    int `mapstructure:"grpc_port"`
	MetricsPort int `mapstructure:"metrics_port"`
	DB          DBConfig
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("user.config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
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
	if cfg.MetricsPort == 0 {
		cfg.MetricsPort = v.GetInt("METRICS_PORT")
	}

	if cfg.GRPCPort == 0 {
		return nil, fmt.Errorf("GRPC_PORT is required")
	}
	if cfg.MetricsPort == 0 {
		return nil, fmt.Errorf("METRICS_PORT is required")
	}
	if cfg.DB.Host == "" {
		return nil, fmt.Errorf("DB_HOST is required")
	}
	if cfg.DB.Port == 0 {
		return nil, fmt.Errorf("DB_PORT is required")
	}
	if cfg.DB.User == "" {
		return nil, fmt.Errorf("DB_USER is required")
	}
	if cfg.DB.DBName == "" {
		return nil, fmt.Errorf("DB_NAME is required")
	}

	return &cfg, nil
}
