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

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type Config struct {
	GRPCPort int `mapstructure:"grpc_port"`
	DB       DBConfig
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

	cfg.Redis.Host = v.GetString("REDIS_HOST")
	cfg.Redis.Port = v.GetInt("REDIS_PORT")
	cfg.Redis.Password = v.GetString("REDIS_PASSWORD")
	cfg.Redis.DB = v.GetInt("REDIS_DB")

	if cfg.GRPCPort == 0 {
		cfg.GRPCPort = v.GetInt("GRPC_PORT")
	}

	if cfg.CSRF.SecretKey == "" {
		cfg.CSRF.SecretKey = v.GetString("CSRF_SECRET_KEY")

		if cfg.CSRF.SecretKey == "" {
			return nil, fmt.Errorf("CSRF_SECRET_KEY is required")
		}
	}
	if cfg.DB.Host == "" {
		return nil, fmt.Errorf("DB_HOST is a required")
	}

	return &cfg, nil
}
