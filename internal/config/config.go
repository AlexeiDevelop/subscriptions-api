package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type DB struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

type Server struct {
	Port int `mapstructure:"port"`
}

type Config struct {
	Env    string `mapstructure:"env"`
	Server Server `mapstructure:"server"`
	DB     DB     `mapstructure:"db"`
}

func Load() (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("env", "dev")
	v.SetDefault("server.port", 8080)
	v.SetDefault("db.host", "localhost")
	v.SetDefault("db.port", 5432)
	v.SetDefault("db.user", "postgres")
	v.SetDefault("db.password", "postgres")
	v.SetDefault("db.name", "subscriptions")
	v.SetDefault("db.sslmode", "disable")

	// YAML
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("configs")
	_ = v.ReadInConfig() // optional

	// ENV override
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	// map env -> keys
	bindEnv := map[string]string{
		"env": "APP_ENV",
		"server.port": "APP_PORT",
		"db.host": "APP_DB_HOST",
		"db.port": "APP_DB_PORT",
		"db.user": "APP_DB_USER",
		"db.password": "APP_DB_PASSWORD",
		"db.name": "APP_DB_NAME",
		"db.sslmode": "APP_DB_SSLMODE",
	}
	for k, e := range bindEnv {
		_ = v.BindEnv(k, e)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config unmarshal: %w", err)
	}

	return &cfg, nil
}

func (db DB) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.Name, db.SSLMode)
}

func IsDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil { return true }
	return false
}
