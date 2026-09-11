package main

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseConfig
}

func Load() (*Config, error) {
	config := Config{}

	if err := godotenv.Load(".env"); err != nil {
		return &Config{}, err
	}

	config.DatabaseConfig.Name = os.Getenv("DB_NAME")
	if config.DatabaseConfig.Name == "" {
		return &Config{}, errors.New("database Name cannot be empty")
	}

	config.DatabaseConfig.User = os.Getenv("DB_USER")
	if config.DatabaseConfig.User == "" {
		return &Config{}, errors.New("database user cannot be empty")
	}

	config.DatabaseConfig.Password = os.Getenv("DB_PASSWORD")
	if config.DatabaseConfig.Password == "" {
		return &Config{}, errors.New("database Password cannot be empty")
	}

	config.DatabaseConfig.Host = os.Getenv("DB_HOST")
	if config.DatabaseConfig.Host == "" {
		return &Config{}, errors.New("database Host cannot be empty")
	}

	config.DatabaseConfig.Port = os.Getenv("DB_PORT")
	if config.DatabaseConfig.Port == "" {
		return &Config{}, errors.New("database Port cannot be empty")
	}

	config.DatabaseConfig.SSLMode = os.Getenv("DB_SSLMODE")
	if config.DatabaseConfig.SSLMode == "" {
		return &Config{}, errors.New("database SSLMode cannot be empty")
	}

	min, err := strconv.Atoi(os.Getenv("DB_MIN_CONNS"))
	if err != nil {
		return &Config{}, errors.New("database Minimal Connections cannot be empty")
	}
	config.DatabaseConfig.MinConns = min

	max, err := strconv.Atoi(os.Getenv("DB_MAX_CONNS"))
	if err != nil {
		return &Config{}, errors.New("database Maximal Connections cannot be empty")
	}
	config.DatabaseConfig.MaxConns = max

	return &config, nil
}
