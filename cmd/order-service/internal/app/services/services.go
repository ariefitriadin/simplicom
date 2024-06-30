package services

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	pgrepo "github.com/ariefitriadin/simplicom/cmd/order-service/internal/persistence/postgres/repositories"

	"github.com/ariefitriadin/simplicom/cmd/order-service/internal/app/config"
	"github.com/ariefitriadin/simplicom/pkg/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

type ServiceContainer struct {
	DB               *sql.DB
	SQL              *pgxpool.Pool
	PersistenceQuery *pgrepo.Queries
	AuthConn         *grpc.ClientConn
}

func NewServiceContainer(ctx context.Context, cfg *config.Config) (*ServiceContainer, error) {
	db := postgres.NewConnection(ctx, postgres.ConnectionConfig{
		Host:     cfg.POSTGRES.Host,
		Port:     cfg.POSTGRES.Port,
		User:     cfg.POSTGRES.User,
		Pass:     cfg.POSTGRES.Pass,
		Database: cfg.POSTGRES.Database,
	})

	return &ServiceContainer{
		SQL:              db,
		PersistenceQuery: pgrepo.New(db),
	}, nil
}

func (sc *ServiceContainer) Close() error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	var errs []error
	go func() {
		defer wg.Done()
		if sc.SQL != nil {
			sc.SQL.Close()
		}
	}()

	go func() {
		defer wg.Done()
		if sc.AuthConn != nil {
			if err := sc.AuthConn.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}()

	wg.Wait()

	var closeErr error
	for _, err := range errs {
		if closeErr == nil {
			closeErr = err
		} else {
			closeErr = fmt.Errorf("%v | %v", closeErr, err)
		}
	}

	if closeErr != nil {
		return closeErr
	}

	return ctx.Err()
}
