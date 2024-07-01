package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ariefitriadin/simplicom/cmd/order-service/internal/app/config"
	"github.com/ariefitriadin/simplicom/cmd/order-service/internal/app/services"
	"github.com/ariefitriadin/simplicom/pkg/application"
	"github.com/ariefitriadin/simplicom/pkg/buildinfo"
	grpcutils "github.com/ariefitriadin/simplicom/pkg/grpc"
	"github.com/ariefitriadin/simplicom/pkg/grpc/middleware"
	httputils "github.com/ariefitriadin/simplicom/pkg/http"
	"github.com/vardius/gocontainer"
	"golang.org/x/exp/rand"
	"google.golang.org/grpc"

	ordergrpc "github.com/ariefitriadin/simplicom/cmd/order-service/internal/interfaces/grpc"
	orderhttp "github.com/ariefitriadin/simplicom/cmd/order-service/internal/interfaces/http"
	proto "github.com/ariefitriadin/simplicom/cmd/order-service/proto"
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
		panic(fmt.Errorf("failed to create product service container: %w", err))
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

	orderGrpcServer := ordergrpc.NewServer(container.PersistenceQuery, container.SQL)

	proto.RegisterOrderServiceServer(grpcServer, orderGrpcServer)

	router := orderhttp.NewRouter(
		container.DB,
		map[string]*grpc.ClientConn{
			"auth": container.AuthConn,
		},
	)

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
			"gRPC order server",
			fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port),
			grpcServer,
		),
	)

	app.WithShutdownTimeout(cfg.App.ShutdownTimeout)
	app.Run(ctx)

}
