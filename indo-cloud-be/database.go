package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseConfig struct {
	User           string
	Password       string
	Host           string
	Port           string
	Name           string
	SSLMode        string
	MinConns       int
	MaxConns       int
	MigrateUp      bool
	MigrationsPath string
}

type DB struct {
	*pgxpool.Pool
}

func NewPostgreSQLDB(cfg *DatabaseConfig) (*DB, error) {
	if cfg.MigrateUp {
		if err := runMigrations(cfg); err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return &DB{}, err
	}

	config.MinConns = int32(cfg.MinConns)
	config.MaxConns = int32(cfg.MaxConns)

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

func runMigrations(cfg *DatabaseConfig) error {
	migrationsPath := cfg.MigrationsPath
	if migrationsPath == "" {
		migrationsPath = "database/migration"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)

	if u, err := url.Parse(dsn); err == nil {
		dsn = u.String()
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dsn,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("Database schema is up to date, no migrations applied")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	fmt.Println("Database migrations applied successfully")
	return nil
}
