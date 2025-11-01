package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type CorsConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type CookieConfig struct {
	SessionDuration int `mapstructure:"session_duration"`
}

type MinioConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Secure    bool   `mapstructure:"secure"`
}

type Postgres struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type Config struct {
	Storages Storages `mapstructure:"storages"`
	Auth     Auth     `mapstructure:"auth"`
}

type Storages struct {
	Minio MinioConfig `mapstructure:"minio"`
	Db    Postgres    `mapstructure:"db"`
}

type Auth struct {
	Cors   CorsConfig   `mapstructure:"cors"`
	Cookie CookieConfig `mapstructure:"cookie"`
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

	overrideFromEnv(&config)

	return &config, nil
}

func overrideFromEnv(cfg *Config) {
	if user := os.Getenv("DB_ROOT_USER"); user != "" {
		cfg.Storages.Db.User = user
	}
	if password := os.Getenv("DB_ROOT_PASSWORD"); password != "" {
		cfg.Storages.Db.Password = password
	}
	if dbname := os.Getenv("DB_NAME"); dbname != "" {
		cfg.Storages.Db.DBName = dbname
	}
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Storages.Db.Host = host
	}
	if portStr := os.Getenv("DB_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Storages.Db.Port = port
		}
	}

	if accessKey := os.Getenv("MINIO_ROOT_USER"); accessKey != "" {
		cfg.Storages.Minio.AccessKey = accessKey
	}
	if secretKey := os.Getenv("MINIO_ROOT_PASSWORD"); secretKey != "" {
		cfg.Storages.Minio.SecretKey = secretKey
	}
	if endpoint := os.Getenv("MINIO_ENDPOINT"); endpoint != "" {
		cfg.Storages.Minio.Endpoint = endpoint
	}
}
