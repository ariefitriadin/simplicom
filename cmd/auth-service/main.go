package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ariefitriadin/simplicom/cmd/auth-service/internal/app/config"

	"github.com/ariefitriadin/simplicom/pkg/application"
	httputils "github.com/ariefitriadin/simplicom/pkg/http"
	"github.com/ariefitriadin/simplicom/pkg/postgres"

	"github.com/ariefitriadin/simplicom/cmd/auth-service/internal/app/services"
	"github.com/ariefitriadin/simplicom/cmd/auth-service/internal/app/services/oauth2"
	authgrpc "github.com/ariefitriadin/simplicom/cmd/auth-service/internal/interfaces/grpc"
	authhttp "github.com/ariefitriadin/simplicom/cmd/auth-service/internal/interfaces/http"
	authproto "github.com/ariefitriadin/simplicom/cmd/auth-service/proto"
	"github.com/ariefitriadin/simplicom/pkg/buildinfo"
	grpcutils "github.com/ariefitriadin/simplicom/pkg/grpc"
	"github.com/ariefitriadin/simplicom/pkg/grpc/middleware"
	"github.com/vardius/gocontainer"
	"google.golang.org/grpc"

	"golang.org/x/exp/rand"
)

func init() {
	rand.Seed(uint64(time.Now().UnixNano()))
	gocontainer.GlobalContainer = nil
}

func main() {
	buildinfo.PrintVersionOrContinue()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.FromEnv()
	container, err := services.NewServiceContainer(ctx, cfg)
	if err != nil {
		panic(fmt.Errorf("failed to create service container: %w", err))
	}

	grpcServer := grpcutils.NewServer(
		grpcutils.ServerConfig{
			ServerMinTime: cfg.GRPC.ServerMinTime,
			ServerTime:    cfg.GRPC.ServerTime,
			ServerTimeout: cfg.GRPC.ServerTimeout,
		},
		[]grpc.UnaryServerInterceptor{
			middleware.TransformUnaryOutgoingError(),
			middleware.CountIncomingUnaryRequests(),
		},
		[]grpc.StreamServerInterceptor{
			middleware.TransformStreamOutgoingError(),
			middleware.CountIncomingStreamRequests(),
		},
	)

	oauth2Server := oauth2.InitServer(cfg, container.OAuth2Manager, cfg.OAuth.InitTimeout, container.PersistenceQuery)

	grpcAuthServer := authgrpc.NewServer(oauth2Server, postgres.NewConnection(ctx, postgres.ConnectionConfig{
		Host:     cfg.POSTGRES.Host,
		Port:     cfg.POSTGRES.Port,
		User:     cfg.POSTGRES.User,
		Pass:     cfg.POSTGRES.Pass,
		Database: cfg.POSTGRES.Database,
	}))

	router := authhttp.NewRouter(
		container.DB,
		map[string]*grpc.ClientConn{
			"auth": container.AuthConn,
		},
	)

	authproto.RegisterAuthServiceServer(grpcServer, grpcAuthServer)

	app := application.New()
	app.AddAdapters(
		httputils.NewAdapter(
			&http.Server{
				Addr:         fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
				ReadTimeout:  cfg.HTTP.ReadTimeout,
				WriteTimeout: cfg.HTTP.WriteTimeout,
				IdleTimeout:  cfg.HTTP.IdleTimeout, // limits server-side the amount of time a Keep-Alive connection will be kept idle before being reused
				Handler:      router,
			},
		),
		grpcutils.NewAdapter(
			"gRPC auth server",
			fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port),
			grpcServer,
		),
	)

	if cfg.App.Environment == "development" {
		app.AddAdapters(
			application.NewDebugAdapter(
				fmt.Sprintf("%s:%d", cfg.Debug.Host, cfg.Debug.Port),
			),
		)
	}

	app.WithShutdownTimeout(cfg.App.ShutdownTimeout)
	app.Run(ctx)
}
