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

	cfg.Minio.Endpoint = v.GetString("MINIO_ENDPOINT")
	cfg.Minio.AccessKey = v.GetString("MINIO_ACCESS_KEY")
	cfg.Minio.SecretKey = v.GetString("MINIO_SECRET_KEY")
	cfg.Minio.Secure = v.GetBool("MINIO_SECURE")

	if host := v.GetString("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if port := v.GetInt("SERVER_PORT"); port != 0 {
		cfg.Server.Port = port
	}

	if key := v.GetString("CSRF_SECRET_KEY"); key != "" {
		cfg.CSRF.SecretKey = key
	}

	if cfg.Services == nil {
		cfg.Services = make(map[string]ServiceConfig)
	}

	serviceNames := []string{"auth", "user", "notes"}
	for _, svc := range serviceNames {
		current, exists := cfg.Services[svc]
		if !exists {
			current = ServiceConfig{}
		}

		hostKey := fmt.Sprintf("SERVICES_%s_GRPC_HOST", strings.ToUpper(svc))
		portKey := fmt.Sprintf("SERVICES_%s_GRPC_PORT", strings.ToUpper(svc))

		if val := v.GetString(hostKey); val != "" {
			current.GrpcHost = val
		}
		if val := v.GetInt(portKey); val != 0 {
			current.GrpcPort = val
		}

		cfg.Services[svc] = current
	}

	if cfg.Server.Port == 0 {
		return nil, fmt.Errorf("SERVER_PORT is required")
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

	for _, svc := range serviceNames {
		sCfg, ok := cfg.Services[svc]
		if !ok || sCfg.GrpcHost == "" || sCfg.GrpcPort == 0 {
			return nil, fmt.Errorf("config for service '%s' is incomplete (host or port missing)", svc)
		}
	}

	return &cfg, nil
}
