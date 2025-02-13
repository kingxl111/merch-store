package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kingxl111/merch-store/internal/config"
	"github.com/kingxl111/merch-store/internal/repository/postgres"
	"github.com/kingxl111/merch-store/internal/shop/service"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := setupLogger()

	if err := runMain(ctx, logger); err != nil {
		logger.Error("run main", slog.Any("err", err))
		os.Exit(1)
	}
}

func setupLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

func runMain(ctx context.Context, logger *slog.Logger) error {
	pgConfig, err := config.NewPGConfig()
	if err != nil {
		return fmt.Errorf("pg config: %w", err)
	}

	db, err := postgres.NewDB(
		pgConfig.Username,
		pgConfig.Password,
		pgConfig.Host,
		pgConfig.Port,
		pgConfig.DBName,
		pgConfig.SSLMode,
	)
	if err != nil {
		return fmt.Errorf("db init: %w", err)
	}
	defer db.Close()

	repo := postgres.NewRepository(db)
	srv := service.NewService(repo)
	handler := http.NewHandler(srv)

	e := echo.New()
	e.GET("/http-server/info", handler.GetApiInfo)

	httpServer := environment.NewServer(e, environment.ServerOptions{
		Logger: logger,
	})

	// Запуск сервера
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return httpServer.ListenAndServeContext(ctx, ":8080")
	})

	eg.Go(func() error {
		<-ctx.Done()
		return httpServer.Shutdown(context.Background())
	})

	return eg.Wait()
}
