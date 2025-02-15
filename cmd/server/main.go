package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

	"github.com/kingxl111/merch-store/internal/config"
	env "github.com/kingxl111/merch-store/internal/environment"
	httpserver "github.com/kingxl111/merch-store/internal/gates/http-server"
	"github.com/kingxl111/merch-store/internal/repository/postgres"
	shop "github.com/kingxl111/merch-store/internal/shop/service"
	usrs "github.com/kingxl111/merch-store/internal/users/service"
	merchstoreapi "github.com/kingxl111/merch-store/pkg/api/merch-store"
)

const baseURL = ""

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

	err := config.Load(configPath)
	if err != nil {
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
	e := echo.New()
	merchstoreapi.RegisterHandlersWithBaseURL(e, handler, baseURL)
	httpServer := opts.NewServer(e, httpServerConfig.Address())

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(
		func() error {
			logger.Info("starting http server on " + httpServerConfig.Address() + "...")
			return env.ListenAndServeContext(ctx, httpServer)
		},
	)

	eg.Go(func() error {
		<-ctx.Done()
		return httpServer.Shutdown(context.Background())
	})

	return eg.Wait()
}
