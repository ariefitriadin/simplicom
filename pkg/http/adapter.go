package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/ariefitriadin/simplicom/pkg/logger"
	"net"
	"net/http"
)

// Adapter is http server app adapter
type Adapter struct {
	httpServer *http.Server
}

// NewAdapter provides new primary HTTP adapter
func NewAdapter(httpServer *http.Server) *Adapter {
	return &Adapter{
		httpServer: httpServer,
	}
}

// Start start http application adapter
func (adapter *Adapter) Start(ctx context.Context) error {
	adapter.httpServer.BaseContext = func(_ net.Listener) context.Context {
		logger.Info(ctx, fmt.Sprintf("http server listening on address :  %s", adapter.httpServer.Addr))
		return ctx
	}

	if err := adapter.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// Stop stops http application adapter
func (adapter *Adapter) Stop(ctx context.Context) error {
	return adapter.httpServer.Shutdown(ctx)
}
