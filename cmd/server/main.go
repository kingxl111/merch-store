package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/kingxl111/merch-store/internal/config"
	env "github.com/kingxl111/merch-store/internal/environment"
	httpserver "github.com/kingxl111/merch-store/internal/gates/http-server"
	"github.com/kingxl111/merch-store/internal/repository/postgres"
	shop "github.com/kingxl111/merch-store/internal/shop/service"
	usrs "github.com/kingxl111/merch-store/internal/users/service"
	merchstoreapi "github.com/kingxl111/merch-store/pkg/api/merch-store"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", ".env", "path to config file")
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	defaultLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(defaultLogger)

	if err := runMain(ctx); err != nil {
		defaultLogger.Error("run main", slog.Any("err", err))
		return
	}
}

func runMain(ctx context.Context) error {
	flag.Parse()

	if err := config.Load(configPath); err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

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

	loggerConfig, err := config.NewLoggerConfig()
	if err != nil {
		return fmt.Errorf("failed to get logger config: %v", err)
	}

	handleOpts := &slog.HandlerOptions{
		Level: loggerConfig.Level(),
	}
	var h slog.Handler = slog.NewTextHandler(os.Stdout, handleOpts)
	logger := slog.New(h)

	repo := postgres.NewRepository(db)
	shopSrv := shop.NewShopService(repo)
	userSrv := usrs.NewUserService(repo, repo)

	httpServerConfig, err := config.NewHTTPConfig()
	if err != nil {
		return fmt.Errorf("http server config error: %w", err)
	}

	var opts env.ServerOptions
	opts.WithLogger(logger)
	handler := httpserver.NewHandler(userSrv, shopSrv)
	mux := http.NewServeMux()
	apiHandler := merchstoreapi.HandlerFromMux(handler, mux)
	httpServer := opts.NewServer(apiHandler, httpServerConfig.Address())

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		logger.Info("starting http server on " + httpServerConfig.Address() + "...")
		if err := env.ListenAndServeContext(ctx, httpServer); err != nil {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()
		logger.Info("shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("error during server shutdown", slog.Any("err", err))
			return err
		}

		logger.Info("server stopped gracefully")
		return nil
	})

	if err := eg.Wait(); err != nil {
		logger.Error("server terminated with error", slog.Any("err", err))
		return fmt.Errorf("server terminated with error: %w", err)
	}

	logger.Info("server exited cleanly")
	return nil
}
