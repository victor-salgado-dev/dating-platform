// Comando api arranca el backend monolítico modular de la plataforma
// de dating. En esta fase (Fase 1) solo expone los endpoints de
// infraestructura necesarios para verificar que el stack funciona:
// liveness/readiness. Los módulos de dominio (auth, users, profiles,
// search, messages, ...) se irán añadiendo en fases posteriores.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dating-platform/backend/internal/config"
	"dating-platform/backend/internal/db"
	"dating-platform/backend/internal/redisclient"
	"dating-platform/backend/internal/server"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.Postgres)
	if err != nil {
		slog.Error("no se pudo conectar a PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	rdb, err := redisclient.New(ctx, cfg.Redis)
	if err != nil {
		slog.Error("no se pudo conectar a Redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()

	router := server.NewRouter(pool, rdb)

	addr := cfg.Backend.Host + ":" + cfg.Backend.Port
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("backend escuchando", "addr", addr, "env", cfg.AppEnv)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error en el servidor HTTP", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("apagando servidor...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("error durante el apagado del servidor", "error", err)
	}

	slog.Info("servidor detenido correctamente")
}
