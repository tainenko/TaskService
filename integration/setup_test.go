//go:build integration

// Package integration runs the HTTP API against a real PostgreSQL container.
// Run with: go test -tags integration ./integration/...
package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github/TaskService/auth"
	"github/TaskService/dao"
	"github/TaskService/handler"
	"github/TaskService/middleware"
	"github/TaskService/router"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	server *httptest.Server
	rawDB  *sql.DB
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()

	pg, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("taskservice"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		log.Printf("failed to start postgres container (is Docker running?): %v", err)
		return 1
	}
	defer func() { _ = testcontainers.TerminateContainer(pg) }()

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("connection string: %v", err)
		return 1
	}

	// Apply the real migrations so schema drift is caught here.
	rawDB, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Printf("open sql db: %v", err)
		return 1
	}
	defer func() { _ = rawDB.Close() }()
	if err := goose.SetDialect("postgres"); err != nil {
		log.Printf("goose dialect: %v", err)
		return 1
	}
	goose.SetLogger(goose.NopLogger())
	if err := goose.Up(rawDB, "../migration"); err != nil {
		log.Printf("migrations failed: %v", err)
		return 1
	}

	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Printf("gorm open: %v", err)
		return 1
	}
	dao.SetDefault(gdb)

	tokens, err := auth.NewTokenManager("integration-test-secret-0123456789abcdef", time.Hour)
	if err != nil {
		log.Printf("token manager: %v", err)
		return 1
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CustomRecovery())
	health := handler.NewHealthHandler(rawDB)
	r.GET("/healthz", health.Healthz)
	r.GET("/readyz", health.Readyz)
	router.SetAuthRoute(r, gdb, tokens)
	router.SetTaskRoute(r, tokens)

	server = httptest.NewServer(r)
	defer server.Close()

	return m.Run()
}

// uniqueEmail keeps tests independent of each other on the shared database.
func uniqueEmail(t *testing.T) string {
	return fmt.Sprintf("%s-%d@example.com", t.Name(), time.Now().UnixNano())
}
