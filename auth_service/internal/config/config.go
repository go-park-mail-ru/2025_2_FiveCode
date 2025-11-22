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
	Host     string
	Port     int
	Password string
	DB       int
}

type ServiceConfig struct {
	GrpcHost string `mapstructure:"grpc_host"`
	GrpcPort int    `mapstructure:"grpc_port"`
}

type Config struct {
	GRPCPort int `mapstructure:"grpc_port"`
	Redis    RedisConfig
	Services map[string]ServiceConfig `mapstructure:"services"`
	Cors     CorsConfig               `mapstructure:"cors"`
	Cookie   CookieConfig             `mapstructure:"cookie"`
	CSRF     CSRFConfig               `mapstructure:"csrf"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("auth.config")
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

	cfg.Redis.Host = v.GetString("REDIS_HOST")
	cfg.Redis.Port = v.GetInt("REDIS_PORT")
	cfg.Redis.Password = v.GetString("REDIS_PASSWORD")
	cfg.Redis.DB = v.GetInt("REDIS_DB")

	if cfg.Services == nil {
		cfg.Services = make(map[string]ServiceConfig)
	}

	userService := cfg.Services["user"]

	if host := v.GetString("SERVICES_USER_GRPC_HOST"); host != "" {
		userService.GrpcHost = host
	}
	if port := v.GetInt("SERVICES_USER_GRPC_PORT"); port != 0 {
		userService.GrpcPort = port
	}

	cfg.Services["user"] = userService

	if cfg.GRPCPort == 0 {
		cfg.GRPCPort = v.GetInt("GRPC_PORT")
	}

	if cfg.CSRF.SecretKey == "" {
		cfg.CSRF.SecretKey = v.GetString("CSRF_SECRET_KEY")

		if cfg.CSRF.SecretKey == "" {
			return nil, fmt.Errorf("CSRF_SECRET_KEY is required")
		}
	}

	return &cfg, nil
}
