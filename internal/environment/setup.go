package environment

import (
	"context"
	"fmt"
	"github.com/kingxl111/merch-store/internal/config"
	"log/slog"
	"strings"

	"errors"
	"github.com/sethvargo/go-envconfig"
)

type closer func()

type Env struct {
	Config  *config.Config
	Logger  *slog.Logger
	Servers *Servers

	Closers []closer
}

func Setup(ctx context.Context) (*Env, error) {
	var cfg config.Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, fmt.Errorf("env processing: %w", err)
	}

	if err := provideTenantID(cfg); err != nil {
		return nil, fmt.Errorf("provideTenantID: %w", err)
	}

	var e Env

	sentryClsr, err := initSentry(cfg)
	if err != nil {
		return nil, fmt.Errorf("initSentry: %w", err)
	}

	logger, err := initLogger(cfg)
	if err != nil {
		return nil, fmt.Errorf("initLogger: %w", err)
	}

	clients, err := newClients(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("newClients: %w", err)
	}

	if err := initializeTopics(ctx, cfg, logger, clients.Broker.Internal); err != nil {
		return nil, fmt.Errorf("initializeTopics: %w", err)
	}

	services, err := newServices(ctx, clients, &cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("newServices: %w", err)
	}

	servers := newServers(ctx, cfg, logger, clients)

	e.Closers = append(e.Closers, sentryClsr, clients.Broker.Internal.Close)

	e.Servers = servers
	e.Config = &cfg
	e.Logger = logger
	e.Clients = clients
	e.Services = services
	return &e, nil
}
