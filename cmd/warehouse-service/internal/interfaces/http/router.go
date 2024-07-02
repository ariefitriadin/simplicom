package http

import (
	"database/sql"
	"net/http"

	"github.com/ariefitriadin/simplicom/cmd/warehouse-service/internal/interfaces/http/handlers"
	"github.com/ariefitriadin/simplicom/pkg/http/response/json"
	"github.com/vardius/gorouter/v4"
	"google.golang.org/grpc"
)

func NewRouter(
	sqlConn *sql.DB,
	grpcConnectionMap map[string]*grpc.ClientConn,
) http.Handler {
	mainRouter := gorouter.New()
	mainRouter.NotFound(json.NotFound())
	mainRouter.NotAllowed(json.NotAllowed())
	// Liveness probes are to indicate that your application is running
	mainRouter.GET("/health", handlers.BuildLivenessHandler())
	// Readiness is meant to check if your application is ready to serve traffic
	mainRouter.GET("/readiness", handlers.BuildReadinessHandler(sqlConn, grpcConnectionMap))

	return mainRouter
}
