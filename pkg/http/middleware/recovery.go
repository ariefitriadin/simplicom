package middleware

import (
	"fmt"
	"github.com/vardius/gorouter/v4"
	"net/http"
	"runtime/debug"

	"github.com/ariefitriadin/simplicom/pkg/logger"

	apperrors "github.com/ariefitriadin/simplicom/pkg/errors"
	"github.com/ariefitriadin/simplicom/pkg/http/response/json"
)

// Recover middleware recovers from panic
func Recover() gorouter.MiddlewareFunc {
	m := func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Critical(r.Context(), fmt.Sprintf("[HTTP] Recovered in %v %s", rec, debug.Stack()))

					appErr := apperrors.Wrap(fmt.Errorf("%w: recovered from panic", apperrors.ErrInternal))

					if err := json.JSONError(r.Context(), w, appErr); err != nil {
						logger.Critical(r.Context(), fmt.Sprintf("[HTTP] Errors while sending response after panic %v", err))
					}
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}

	return m
}
