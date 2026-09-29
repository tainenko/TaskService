package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github/TaskService/auth"
	"github/TaskService/conf"
	"github/TaskService/dao"
	"github/TaskService/handler"
	"github/TaskService/metrics"
	"github/TaskService/middleware"
	"github/TaskService/ratelimit"
	"github/TaskService/router"
	"github/TaskService/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const shutdownTimeout = 10 * time.Second

func main() {
	env := flag.String("env", envOr("APP_ENV", "local"), "Environment: local|dev|prod")
	flag.Parse()

	if err := run(*env); err != nil {
		slog.Error("service exited with error", "error", err)
		os.Exit(1)
	}
}

func run(env string) error {
	config, err := conf.LoadConfig(env)
	if err != nil {
		return err
	}
	slog.Info("config loaded", "env", env, "mode", config.Server.RunMode, "port", config.Server.HttpPort)

	db, sqlDB, err := setupDB(config.Database)
	if err != nil {
		return err
	}
	defer func() { _ = sqlDB.Close() }()

	tokens, err := auth.NewTokenManager(config.Auth.JWTSecret, time.Duration(config.Auth.TokenTTLMinutes)*time.Minute)
	if err != nil {
		return fmt.Errorf("auth config: %w (set TASK_AUTH_JWTSECRET)", err)
	}

	if config.Server.RunMode != "" {
		gin.SetMode(config.Server.RunMode)
	}

	// gin.New instead of gin.Default so the built-in Recovery does not shadow ours.
	r := gin.New()
	if err := r.SetTrustedProxies(config.Server.TrustedProxies); err != nil {
		return fmt.Errorf("invalid trusted proxies: %w", err)
	}
	m := metrics.New(sqlDB)
	r.Use(gin.Logger(), middleware.CustomRecovery(), m.Middleware())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to the API",
		})
	})
	r.GET("/metrics", m.Handler(config.Server.MetricsToken))
	health := handler.NewHealthHandler(sqlDB)
	r.GET("/healthz", health.Healthz)
	r.GET("/readyz", health.Readyz)
	ipLimiter := ratelimit.NewKeyedLimiter(
		ratelimit.PerMinute(config.Auth.IPRatePerMinute), config.Auth.IPBurst, time.Hour)
	// Refill so that LoginMaxFailures failures are forgiven over one window.
	failureWindow := time.Duration(config.Auth.LoginFailureWindowMinutes) * time.Minute
	loginFailures := ratelimit.NewKeyedLimiter(
		ratelimit.PerMinute(float64(config.Auth.LoginMaxFailures)/failureWindow.Minutes()),
		config.Auth.LoginMaxFailures, failureWindow)
	sessions := service.NewRefreshService(db, time.Duration(config.Auth.RefreshTTLHours)*time.Hour)
	router.SetAuthRoute(r, db, tokens, sessions, ipLimiter, loginFailures)
	router.SetTaskRoute(r, tokens)

	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(config.Server.HttpPort),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("shutting down", "timeout", shutdownTimeout)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func setupDB(config conf.Database) (*gorm.DB, *sql.DB, error) {
	// url.URL escapes special characters in the credentials.
	dsn := (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(config.Username, config.Password),
		Host:     config.Host,
		Path:     "/" + config.DBName,
		RawQuery: url.Values{"sslmode": {config.SSLMode}}.Encode(),
	}).String()

	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)

	dao.SetDefault(db)
	return db, sqlDB, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
