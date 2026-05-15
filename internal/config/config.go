package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env       string        `yaml:"env" env-default:"local" env-required:"true"`
	Serv      Server        `yaml:"serv"`
	TokenTTL  time.Duration `yaml:"token_ttl env-default:"1800s"`
	DSN       string
	SecretKey string
}

type Server struct {
	Port          string        `yaml:"port" env-default:":8090"`
	Timeout       time.Duration `yaml:"timeout" env-default:"4s"`
	Idle_timeoout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustParseConfig() *Config {
	configPath := os.Getenv("SSO_CONFIG_PATH")
	if configPath == "" {
		panic("SSO_CONFIG_PATH isn't set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic(fmt.Sprintf("config file does not exist: %s", configPath))
	}

	dsn, err := loadDSN()
	if err != nil {
		panic("failed to load DSN")
	}

	key, err := loadKey()
	if err != nil {
		panic("failed to load secret key")
	}

	cfg := Config{DSN: dsn, SecretKey: key}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(fmt.Sprintf("cannot read config file: %s\n", err))
	}

	return &cfg
}

func loadDSN() (string, error) {
	if err := godotenv.Load(); err != nil {
		return "", err
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	return fmt.Sprintf(
		`user=%s password=%s host=%s port=%s dbname=%s sslmode=allow`,
		user, password, host, port, dbname,
	), nil
}

func loadKey() (string, error) {
	if err := godotenv.Load(); err != nil {
		return "", err
	}

	secretKey := os.Getenv("SECRET_KEY")

	return secretKey, nil
}
