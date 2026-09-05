package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/naizechi/friday/internal/core/logger"
	"github.com/naizechi/friday/internal/core/repository/postgres/pool"
	"github.com/naizechi/friday/internal/core/transport/http/middleware"
	"github.com/naizechi/friday/internal/core/transport/http/server"
	healthhttp "github.com/naizechi/friday/internal/features/health/transport/http"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log, err := logger.NewLogger(logger.NewConfigMust())
	if err != nil {
		panic(err)
	}
	defer log.Close()

	log.Debug("initializing postgres connection pool")
	pgPool, err := pool.NewConnectionPool(ctx, pool.NewConfigMust())
	if err != nil {
		log.Error("failed to init postgres connection pool", "err", err)
		os.Exit(1)
	}
	defer pgPool.Close()

	log.Debug("initializing HTTP server")
	httpServer := server.NewHTTPServer(
		server.NewConfigMust(),
		log,
		middleware.RequestID(),
		middleware.Logger(log),
		middleware.Recoverer(),
		middleware.Trace(),
	)

	healthTransportHTTP := healthhttp.NewHandler()

	httpServer.RegisterAPIRoutes("v1", func(r chi.Router) {
		healthTransportHTTP.Register(r)
	})

	if err := httpServer.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("HTTP server run error", "err", err)
		os.Exit(1)
	}
}
