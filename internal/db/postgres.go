package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

// Init initializes the PostgreSQL connection and creates the secrets table.
func Init() error {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	table := os.Getenv("DB_TABLE")

	if user == "" || password == "" || host == "" || dbName == "" {
		return fmt.Errorf("missing required DB environment variables")
	}

	if port == "" {
		port = "5432"
	}
	if table == "" {
		table = "secrets"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbName)

	var err error
	DB, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}

	createTable := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id UUID PRIMARY KEY,
		encrypted_data BYTEA NOT NULL,
		iv BYTEA NOT NULL,
		expires_at TIMESTAMPTZ NOT NULL,
		viewed BOOLEAN NOT NULL DEFAULT false
	);`, table)

	_, err = DB.Exec(context.Background(), createTable)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}
