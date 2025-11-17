package config

import (
	"fmt"

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

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Secure    bool
}

type PostgresConfig struct {
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
	Storages Storages `mapstructure:"storages"`
	Auth     Auth     `mapstructure:"auth"`
}

type Storages struct {
	Minio MinioConfig    `mapstructure:"minio"`
	Db    PostgresConfig `mapstructure:"db"`
	Redis RedisConfig    `mapstructure:"redis"`
}

type Auth struct {
	Cors   CorsConfig   `mapstructure:"cors"`
	Cookie CookieConfig `mapstructure:"cookie"`
	CSRF   CSRFConfig   `mapstructure:"csrf"`
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

	if err := loadEnvFile(); err != nil {
		return nil, fmt.Errorf("failed to load env file: %w", err)
	}

	getMinioCfg(&config)
	getDbCfg(&config)
	getRedisCfg(&config)

	return &config, nil
}

func getDbCfg(c *Config) {
	c.Storages.Db = PostgresConfig{
		Host:     viper.GetString("DB_HOST"),
		Port:     viper.GetInt("DB_PORT"),
		User:     viper.GetString("DB_USER"),
		Password: viper.GetString("DB_PASSWORD"),
		DBName:   viper.GetString("DB_NAME"),
		SSLMode:  viper.GetString("DB_SSLMODE"),
	}
}

func getMinioCfg(c *Config) {
	c.Storages.Minio = MinioConfig{
		Endpoint:  viper.GetString("MINIO_ENDPOINT"),
		AccessKey: viper.GetString("MINIO_ACCESS_KEY"),
		SecretKey: viper.GetString("MINIO_SECRET_KEY"),
		Secure:    viper.GetBool("MINIO_SECURE"),
	}
}

func getRedisCfg(c *Config) {
	c.Storages.Redis = RedisConfig{
		Host:     viper.GetString("REDIS_HOST"),
		Port:     viper.GetInt("REDIS_PORT"),
		Password: viper.GetString("REDIS_PASSWORD"),
		DB:       viper.GetInt("REDIS_DB"),
	}
}
