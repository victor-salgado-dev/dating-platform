// Comando api arranca el backend monolítico modular de la plataforma
// de dating. Módulos activos: health, auth, profiles, search,
// favorites, messaging, blocking, reports, admin, consent y contact.
// El resto de módulos de dominio se irán añadiendo en fases futuras
// (producción).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"dating-platform/backend/internal/admin"
	"dating-platform/backend/internal/auth"
	"dating-platform/backend/internal/blocking"
	"dating-platform/backend/internal/config"
	"dating-platform/backend/internal/consent"
	"dating-platform/backend/internal/contact"
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

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(cfg.LogLevel),
	}))
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

	emailSender, err := email.NewSender(cfg.Email.Driver, email.SMTPConfig{
		Host:     cfg.Email.SMTPHost,
		Port:     cfg.Email.SMTPPort,
		Username: cfg.Email.SMTPUsername,
		Password: cfg.Email.SMTPPassword,
		From:     cfg.Email.From,
	})
	if err != nil {
		slog.Error("configuración de email inválida", "error", err)
		os.Exit(1)
	}

	sessionStore := auth.NewRedisSessionStore(rdb, cfg.Auth.SessionTTL)
	tokenStore := auth.NewRedisTokenStore(rdb)

	consentRepo := consent.NewPostgresRepository(pool)
	consentService := consent.NewService(consentRepo)
	consentHandler := consent.NewHandler(consentService)

	authService := auth.NewService(
		usersRepo,
		sessionStore,
		tokenStore,
		emailSender,
		consentService,
		cfg.Auth.SessionTTL,
		cfg.Auth.EmailVerificationTTL,
		cfg.Auth.PasswordResetTTL,
	)
	authHandler := auth.NewHandler(authService, cfg.Auth)

	contactService := contact.NewService(emailSender, cfg.Email.ContactInbox)
	contactHandler := contact.NewHandler(contactService)

	fileStorage, err := storage.New(ctx, storage.Config{
		Driver:    cfg.Storage.Driver,
		LocalPath: cfg.Storage.LocalPath,
		S3: storage.S3Config{
			Endpoint:       cfg.Storage.S3Endpoint,
			Region:         cfg.Storage.S3Region,
			Bucket:         cfg.Storage.S3Bucket,
			AccessKey:      cfg.Storage.S3AccessKey,
			SecretKey:      cfg.Storage.S3SecretKey,
			UseSSL:         cfg.Storage.S3UseSSL,
			ForcePathStyle: cfg.Storage.S3ForcePathStyle,
		},
	})
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
		ConsentHandler:   consentHandler,
		ContactHandler:   contactHandler,
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

// parseLogLevel traduce LOG_LEVEL a slog.Level. Un valor desconocido o
// vacío cae en "info" (el valor por defecto de config.Load), nunca en
// un error de arranque: el nivel de log no debería poder tumbar la app.
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
