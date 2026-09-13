// Comando api arranca el backend monolítico modular de la plataforma
// de dating. Módulos activos: health, auth, profiles, search,
// favorites, messaging, blocking, reports y admin. El resto de módulos
// de dominio se irán añadiendo en fases futuras (seguridad, tests,
// legal, producción).
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

	"dating-platform/backend/internal/admin"
	"dating-platform/backend/internal/auth"
	"dating-platform/backend/internal/blocking"
	"dating-platform/backend/internal/config"
	"dating-platform/backend/internal/db"
	"dating-platform/backend/internal/email"
	"dating-platform/backend/internal/favorites"
	"dating-platform/backend/internal/messaging"
	"dating-platform/backend/internal/profiles"
	"dating-platform/backend/internal/ratelimit"
	"dating-platform/backend/internal/redisclient"
	"dating-platform/backend/internal/reports"
	"dating-platform/backend/internal/search"
	"dating-platform/backend/internal/server"
	"dating-platform/backend/internal/storage"
	"dating-platform/backend/internal/users"
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

	// --- Wiring de módulos de dominio ---------------------------------
	usersRepo := users.NewPostgresRepository(pool)

	emailSender, err := email.NewSender(cfg.Email.Driver)
	if err != nil {
		slog.Error("configuración de email inválida", "error", err)
		os.Exit(1)
	}

	sessionStore := auth.NewRedisSessionStore(rdb, cfg.Auth.SessionTTL)
	tokenStore := auth.NewRedisTokenStore(rdb)

	authService := auth.NewService(
		usersRepo,
		sessionStore,
		tokenStore,
		emailSender,
		cfg.Auth.SessionTTL,
		cfg.Auth.EmailVerificationTTL,
		cfg.Auth.PasswordResetTTL,
	)
	authHandler := auth.NewHandler(authService, cfg.Auth)

	fileStorage, err := storage.New(cfg.Storage.Driver, cfg.Storage.LocalPath)
	if err != nil {
		slog.Error("configuración de storage inválida", "error", err)
		os.Exit(1)
	}

	profilesRepo := profiles.NewPostgresRepository(pool)
	profilesService := profiles.NewService(profilesRepo, fileStorage)
	profilesHandler := profiles.NewHandler(profilesService)

	searchRepo := search.NewPostgresRepository(pool)
	searchService := search.NewService(searchRepo)
	searchHandler := search.NewHandler(searchService)

	favoritesRepo := favorites.NewPostgresRepository(pool)
	favoritesService := favorites.NewService(favoritesRepo, profilesRepo)
	favoritesHandler := favorites.NewHandler(favoritesService)

	blockingRepo := blocking.NewPostgresRepository(pool)
	blockingService := blocking.NewService(blockingRepo, profilesRepo)
	blockingHandler := blocking.NewHandler(blockingService)

	messagingRepo := messaging.NewPostgresRepository(pool)
	messagingService := messaging.NewService(messagingRepo, profilesRepo, blockingRepo)
	messagingHandler := messaging.NewHandler(messagingService)

	reportsRepo := reports.NewPostgresRepository(pool)
	reportsService := reports.NewService(reportsRepo, profilesRepo)
	reportsHandler := reports.NewHandler(reportsService)

	adminService := admin.NewService(usersRepo, reportsRepo)
	adminHandler := admin.NewHandler(adminService)

	rateLimiter := ratelimit.New(rdb)

	router := server.NewRouter(server.Dependencies{
		DB:               pool,
		Redis:            rdb,
		AuthService:      authService,
		AuthHandler:      authHandler,
		SessionCookie:    cfg.Auth.CookieName,
		ProfilesHandler:  profilesHandler,
		SearchHandler:    searchHandler,
		FavoritesHandler: favoritesHandler,
		MessagingHandler: messagingHandler,
		BlockingHandler:  blockingHandler,
		ReportsHandler:   reportsHandler,
		AdminHandler:     adminHandler,
		RateLimiter:      rateLimiter,
		Security:         cfg.Security,
	})

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
