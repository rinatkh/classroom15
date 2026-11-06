package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rinatkh/classroom15/internal/auth"
	"github.com/rinatkh/classroom15/internal/httpapi"
	"github.com/rinatkh/classroom15/internal/httpapi/handler"
	"github.com/rinatkh/classroom15/internal/httpapi/middleware"
	"github.com/rinatkh/classroom15/internal/logging"
	"github.com/rinatkh/classroom15/internal/repository/postgres"
	"github.com/rinatkh/classroom15/internal/task"
	"github.com/rinatkh/classroom15/internal/user"
)

func main() {
	if err := logging.Configure("task-api"); err != nil {
		logging.Fatal("invalid logging configuration", "error", err)
	}
	if err := run(); err != nil {
		logging.Fatal("task-api stopped", "error", err)
	}
}

func run() error {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return errors.New("JWT_SECRET должен содержать не менее 32 байт; см. README")
	}
	accessTTL, err := duration("ACCESS_TOKEN_TTL", "15m")
	if err != nil {
		return err
	}
	refreshTTL, err := duration("REFRESH_TOKEN_TTL", "168h")
	if err != nil {
		return err
	}
	if refreshTTL <= accessTTL {
		return errors.New("REFRESH_TOKEN_TTL должен быть больше ACCESS_TOKEN_TTL")
	}
	secure, err := strconv.ParseBool(envOr("COOKIE_SECURE", "true"))
	if err != nil {
		return errors.New("COOKIE_SECURE должен быть true или false")
	}
	cookie := handler.CookieConfig{Name: envOr("REFRESH_COOKIE_NAME", "refresh_token"), Secure: secure}
	if err := cookie.Validate(); err != nil {
		return errors.New("неверная настройка refresh cookie; см. README")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL обязателен; см. docs/setup.md")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return errors.New("не удалось создать pool PostgreSQL; проверьте DATABASE_URL")
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return errors.New("PostgreSQL недоступен; проверьте подключение и docs/setup.md")
	}
	slog.Info("PostgreSQL connected")

	// User/Task repositories сохраняются; новая зависимость относится только к auth.
	userRepository := postgres.NewUserRepository(pool)
	taskRepository := postgres.NewTaskRepository(pool)
	sessionRepository := postgres.NewRefreshSessionRepository(pool)
	passwordHasher := auth.NewBcryptHasher(0)
	tokenManager := auth.NewJWTManager(secret, accessTTL)
	authUseCase := user.NewUseCase(userRepository, passwordHasher, tokenManager,
		sessionRepository, auth.NewCryptoRefreshTokenGenerator(),
		user.SessionConfig{AccessTTL: accessTTL, RefreshTTL: refreshTTL})
	taskUseCase := task.NewUseCase(taskRepository)
	authHandler := handler.NewAuthHandler(authUseCase, cookie)
	taskHandler := handler.NewTaskHandler(taskUseCase)
	authMiddleware := middleware.NewAuth(tokenManager)
	router := httpapi.NewRouter(authHandler, taskHandler, authMiddleware)
	server := &http.Server{Addr: envOr("HTTP_ADDR", ":8080"), Handler: router, ReadHeaderTimeout: 5 * time.Second}
	slog.Info("HTTP server starting", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func duration(name, fallback string) (time.Duration, error) {
	value, err := time.ParseDuration(envOr(name, fallback))
	if err != nil || value < time.Second || value%time.Second != 0 {
		return 0, fmt.Errorf("%s должен быть duration в целых секундах, не менее 1s", name)
	}
	return value, nil
}
