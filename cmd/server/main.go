package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/samber/lo"
	_ "go.uber.org/automaxprocs"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	defaultLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})).
		With(
			slog.String("build_tag", buildinfo.Tag()),
			slog.String("build_sha", buildinfo.SHA()),
			slog.String("build_time", buildinfo.Time()),
		)
	if err := runMain(ctx); err != nil {
		defaultLogger.Error("run main", slog.Any("err", err))
		return
	}
}

func runMain(ctx context.Context) error {
	env, err := environment.Setup(ctx)
	if err != nil {
		return fmt.Errorf("setup.Setup: %w", err)
	}

	cfg := env.Config
	logger := env.Logger
	observabilityHTTP := env.Servers.HTTP.Observability
	kClientInt := env.Clients.Broker.Internal
	dbClient := env.Clients.MongoDB
	services := env.Services
	APIServerHTTP := env.Servers.HTTP.API

	eg, gctx := errgroup.WithContext(ctx)

	shutdownFuncs := []func(){
		func() {
			if err := observabilityHTTP.Shutdown(gctx); err != nil {
				logger.Error("observability http Shutdown", slog.Any("err", err))
			}
		},
		func() {
			err := dbClient.Disconnect(gctx)
			if err != nil {
				logger.Error("db client shutdown", slog.Any("err", err))
			}
		},
		func() {
			if err := APIServerHTTP.Shutdown(gctx); err != nil {
				logger.Error("api server http Shutdown", slog.Any("err", err))
			}
		},
	}
	for _, f := range env.Closers {
		shutdownFuncs = append(shutdownFuncs, f)
	}
	eg.Go(prepareShutdownCallback(gctx, shutdownFuncs...))

	eg.Go(
		func() error {
			logger.Info(fmt.Sprintf("observability http was started %s", cfg.Observability.ADDR()))
			return observabilityHTTP.ListenAndServe()
		},
	)

	eg.Go(
		func() error {
			p := rpoller.NewReplicationPoller(
				kClientInt,
				logger.WithGroup("kafkaPoller"),
				rpoller.WithSkipRetry(),
				rpoller.WithFetchErrorRetryTimeout(5*time.Second),
				rpoller.WithHandleBackoff(
					func(n int) time.Duration {
						return backoff.ExponentialBackoff(5*time.Second, 30*time.Second, n)
					},
				),
			)
			logger.Info("start create poller")
			lo.ForEach(
				services.Consumers, func(c environment.Consumer, _ int) {
					p.Handle(c.Topic, c.Handler)
				},
			)

			return p.Poll(gctx)
		},
	)

	lo.ForEach(
		env.Services.Workers, func(w worker.Worker, _ int) {
			eg.Go(
				func() error {
					worker.New(logger, w).Start(gctx)
					return nil
				},
			)
		},
	)

	eg.Go(
		func() error {
			logger.Info(fmt.Sprintf("api server http was started %s", cfg.APIServer.ADDR()))
			return APIServerHTTP.ListenAndServe()
		},
	)

	if err := eg.Wait(); err != nil {
		// Do not panic or fatal
		// in order to continue graceful shutdown.
		logger.Error("run main", slog.Any("err", err))
	}

	return nil
}

func prepareShutdownCallback(ctx context.Context, fns ...func()) func() error {
	return func() error {
		<-ctx.Done()

		deadlineCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		closer := make(chan struct{})

		go func() {
			defer cancel()
			defer close(closer)

			for _, fn := range fns {
				fn()
			}
		}()

		select {
		case <-deadlineCtx.Done():
		case <-closer:
		}

		return nil
	}
}
