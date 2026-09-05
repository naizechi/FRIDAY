package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/naizechi/friday/internal/core/logger"
	"github.com/naizechi/friday/internal/core/transport/http/middleware"
)

type HTTPServer struct {
	router chi.Router
	config Config
	log    *logger.Logger
}

func NewHTTPServer(
	config Config,
	log *logger.Logger,
	middleware ...middleware.Middleware,
) *HTTPServer {
	router := chi.NewRouter()

	for _, mw := range middleware {
		router.Use(mw)
	}

	return &HTTPServer{
		router: router,
		config: config,
		log:    log,
	}
}

func (h *HTTPServer) RegisterAPIRoutes(version string, register func(r chi.Router)) {
	h.router.Route("/api/"+version, register)
}

func (h *HTTPServer) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:              h.config.Addr,
		Handler:           h.router,
		ReadHeaderTimeout: h.config.ReadHeaderTimeout,
		ReadTimeout:       h.config.ReadTimeout,
		WriteTimeout:      h.config.WriteTimeout,
		IdleTimeout:       h.config.IdleTimeout,
	}

	ch := make(chan error, 1)

	go func() {
		defer close(ch)

		h.log.Warn("start HTTP server", "addr", h.config.Addr)

		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			ch <- err
		}
	}()

	select {
	case err := <-ch:
		if err != nil {
			return fmt.Errorf("listen and serve HTTP: %w", err)
		}
	case <-ctx.Done():
		h.log.Warn("shutdown HTTP server...")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			h.config.ShutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()

			return fmt.Errorf("shutdown HTTP server: %w", err)
		}

		h.log.Warn("HTTP server stopped")
	}

	return nil
}
