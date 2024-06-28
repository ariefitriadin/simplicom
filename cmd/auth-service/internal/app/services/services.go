package services

import (
	"context"
	"database/sql"
	"fmt"
	pgrepo "github.com/ariefitriadin/simplicom/cmd/auth-service/internal/persistence/postgres/repositories"
	"sync"
	"time"

	"github.com/ariefitriadin/simplicom/cmd/auth-service/internal/app/config"
	"github.com/ariefitriadin/simplicom/pkg/auth"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

type ServiceContainer struct {
	DB               *sql.DB
	SQL              *pgxpool.Pool
	PersistenceQuery *pgrepo.Queries
	AuthConn         *grpc.ClientConn
	OAuth2Manager    oauth2.Manager
	Authenticator    auth.Authenticator
	TokenAuthorizer  auth.TokenAuthorizer
}

func NewServiceContainer(ctx context.Context, cfg *config.Config) (*ServiceContainer, error) {
	return &ServiceContainer{}, nil
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
