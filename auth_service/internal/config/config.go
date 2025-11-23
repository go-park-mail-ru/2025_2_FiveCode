package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type CorsConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type CookieConfig struct {
	SessionDuration int `mapstructure:"session_duration"`
}

type CSRFConfig struct {
	SecretKey       string `mapstructure:"secret_key"`
	TokenTTLMinutes int    `mapstructure:"token_ttl_minutes"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type Config struct {
	GRPCPort int `mapstructure:"grpc_port"`
	Redis    RedisConfig
	Cors     CorsConfig   `mapstructure:"cors"`
	Cookie   CookieConfig `mapstructure:"cookie"`
	CSRF     CSRFConfig   `mapstructure:"csrf"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("auth.config")
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

	if cfg.Redis.Host == "" {
		cfg.Redis.Host = v.GetString("REDIS_HOST")
	}
	if cfg.Redis.Port == 0 {
		cfg.Redis.Port = v.GetInt("REDIS_PORT")
	}
	if cfg.Redis.Password == "" {
		cfg.Redis.Password = v.GetString("REDIS_PASSWORD")
	}

	if cfg.GRPCPort == 0 {
		cfg.GRPCPort = v.GetInt("GRPC_PORT")
	}

	if cfg.CSRF.SecretKey == "" {
		cfg.CSRF.SecretKey = v.GetString("CSRF_SECRET_KEY")
	}

	if cfg.GRPCPort == 0 {
		return nil, fmt.Errorf("GRPC_PORT is required")
	}
	if cfg.Redis.Host == "" {
		return nil, fmt.Errorf("REDIS_HOST is required")
	}
	if cfg.Redis.Port == 0 {
		return nil, fmt.Errorf("REDIS_PORT is required")
	}
	if cfg.CSRF.SecretKey == "" {
		return nil, fmt.Errorf("CSRF_SECRET_KEY is required")
	}

	return &cfg, nil
}
