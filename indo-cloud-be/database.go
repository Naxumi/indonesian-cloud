package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
	SSLMode  string
	MinConns int
	MaxConns int
}

type DB struct {
	*pgxpool.Pool
}

func NewPostgreSQLDB(cfg *DatabaseConfig) (*DB, error) {
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return &DB{}, err
	}

	config.MinConns = int32(cfg.MinConns)
	config.MaxConns = int32(cfg.MinConns)

	conn, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return &DB{}, err
	}
	if err := conn.Ping(context.Background()); err != nil {
		return &DB{}, err
	}

	return &DB{
		conn,
	}, nil
}
