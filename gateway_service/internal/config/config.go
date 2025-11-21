package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

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

type ServiceConfig struct {
	GrpcHost string `mapstructure:"grpc_host"`
	GrpcPort int    `mapstructure:"grpc_port"`
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Secure    bool
}

type Config struct {
	Server   ServerConfig
	DB       DBConfig
	Minio    MinioConfig
	Services map[string]ServiceConfig `mapstructure:"services"`
	Cors     CorsConfig               `mapstructure:"cors"`
	Cookie   CookieConfig             `mapstructure:"cookie"`
	CSRF     CSRFConfig               `mapstructure:"csrf"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("gateway.config")
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

	cfg.Minio.Endpoint = v.GetString("MINIO_ENDPOINT")
	cfg.Minio.AccessKey = v.GetString("MINIO_ACCESS_KEY")
	cfg.Minio.SecretKey = v.GetString("MINIO_SECRET_KEY")
	cfg.Minio.Secure = v.GetBool("MINIO_SECURE")
	
	if cfg.Server.Host == "" {
		cfg.Server.Host = v.GetString("SERVER_HOST")
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = v.GetInt("SERVER_PORT")
	}
	if cfg.CSRF.SecretKey == "" {
		cfg.CSRF.SecretKey = v.GetString("CSRF_SECRET_KEY")
	}
	
	if cfg.CSRF.SecretKey == "" {
		return nil, fmt.Errorf("CSRF_SECRET_KEY is required")
	}
	if cfg.DB.Host == "" {
		return nil, fmt.Errorf("DB_HOST is required")
	}
	if cfg.Minio.Endpoint == "" {
		return nil, fmt.Errorf("MINIO_ENDPOINT is required")
	}

	return &cfg, nil
}