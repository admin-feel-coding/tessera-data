// Package main is the entry point for the tessera-data service.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/feel-coding/tessera-data/internal/config"
	"github.com/feel-coding/tessera-data/internal/handler"
	"github.com/feel-coding/tessera-data/internal/httpx"
	"github.com/feel-coding/tessera-data/internal/store"
)

func main() {
	// Best-effort .env load — ignore error if file does not exist.
	_ = godotenv.Load()

	cfg := config.Load()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	ctx := context.Background()

	pool, err := store.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Warn("could not initialise database pool", "error", err)
	}

	if pool != nil {
		if err := runMigrations(ctx, pool); err != nil {
			slog.Error("migration failed", "error", err)
			os.Exit(1)
		}
	}

	usersH := handler.NewUsers(pool)
	ipH := handler.NewIP(pool)
	devicesH := handler.NewDevices(pool)
	blacklistH := handler.NewBlacklist(pool)
	casesH := handler.NewCases(pool)
	verdictsH := handler.NewVerdicts(pool)

	r := chi.NewRouter()

	// Health is unauthenticated — used by Docker Compose and Cloud Run readiness checks.
	r.Get("/health", handler.Health)

	// All other routes require a valid X-Internal-Key header.
	r.Group(func(r chi.Router) {
		r.Use(httpx.AuthMiddleware(cfg.InternalAPIKey))
		r.Use(httpx.LoggingMiddleware)

		r.Get("/users/{id}/history", usersH.GetHistory)
		r.Get("/ip/{ip}/risk", ipH.GetRisk)
		r.Get("/devices/{id}/fingerprint", devicesH.GetFingerprint)
		r.Get("/blacklist/check", blacklistH.Check)
		r.Post("/cases/similar", casesH.FindSimilar)
		r.Post("/cases", casesH.Save)
		r.Post("/verdicts", verdictsH.Save)
		r.Get("/verdicts", verdictsH.List)
		r.Get("/verdicts/{transaction_id}", verdictsH.GetByTransactionID)
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("tessera-data starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

// runMigrations reads all .sql files from the migrations/ directory (sorted by name)
// and executes each one in a transaction. Safe to re-run because every statement uses IF NOT EXISTS.
func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := os.ReadDir("migrations")
	if err != nil {
		return err
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		sql, err := os.ReadFile(filepath.Join("migrations", name))
		if err != nil {
			return err
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}

		slog.Info("migration applied", "file", name)
	}

	return nil
}
