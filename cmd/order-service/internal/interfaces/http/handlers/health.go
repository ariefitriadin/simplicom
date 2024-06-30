package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	apperrors "github.com/ariefitriadin/simplicom/pkg/errors"
	grpcutils "github.com/ariefitriadin/simplicom/pkg/grpc"
	httpjson "github.com/ariefitriadin/simplicom/pkg/http/response/json"
	"google.golang.org/grpc"
)

// BuildLivenessHandler provides liveness handler
func BuildLivenessHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}

	return http.HandlerFunc(fn)
}

// BuildReadinessHandler provides readiness handler
func BuildReadinessHandler(sqlConn *sql.DB, connMap map[string]*grpc.ClientConn) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) error {
		if sqlConn != nil {
			if err := sqlConn.PingContext(r.Context()); err != nil {
				return apperrors.Wrap(err)
			}
		}

		for name, conn := range connMap {
			if conn == nil {
				// logger.Info(r.Context(), fmt.Sprintf("gRPC connection name %s is nil\n", name))
				continue
			}
			if !grpcutils.IsConnectionServing(r.Context(), name, conn) {
				return apperrors.New(fmt.Sprintf("gRPC connection %s is not serving", name))
			}
		}

		w.WriteHeader(http.StatusNoContent)

		return nil
	}

	return httpjson.HandlerFunc(fn)
}
