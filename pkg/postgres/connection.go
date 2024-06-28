package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/ariefitriadin/simplicom/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConnectionConfig struct {
	Host     string
	Port     int
	User     string
	Pass     string
	Database string
}

func NewConnection(ctx context.Context, cfg ConnectionConfig) *pgxpool.Pool {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.Database)
	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		logger.Critical(ctx, fmt.Sprintf("[POSTGRES|Connection] %v", err))
		os.Exit(1)
	}

	return dbpool
}
