package http

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
)

type ServerOptions struct {
	logger        *slog.Logger
	panicHandler  func(w http.ResponseWriter, r *http.Request, p any)
	middlewares   []func(http.Handler) http.Handler
	serverOptions []func(*http.Server)
}

func (o *ServerOptions) WithLogger(logger *slog.Logger) {
	o.logger = logger
}

func (o *ServerOptions) WithPanicHandler(h func(w http.ResponseWriter, r *http.Request, p any)) {
	o.panicHandler = h
}

func (o *ServerOptions) WithServerOptions(v ...func(*http.Server)) {
	o.serverOptions = append(o.serverOptions, v...)
}

func (o *ServerOptions) WithMiddlewares(v ...func(http.Handler) http.Handler) {
	o.middlewares = append(o.middlewares, v...)
}

func (o *ServerOptions) NewServer(handler http.Handler) *http.Server {
	if o.logger == nil {
		o.logger = slog.Default()
	}

	if o.panicHandler == nil {
		o.panicHandler = func(w http.ResponseWriter, r *http.Request, p any) {
			o.logger.Error("recovered from panic",
				"panic", p,
				"stack", debug.Stack(),
				"method", r.Method,
				"path", r.URL.Path,
			)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}

	wrappedHandler := handler
	for _, mw := range o.middlewares {
		wrappedHandler = mw(wrappedHandler)
	}

	wrappedHandler = o.loggingMiddleware(wrappedHandler)
	wrappedHandler = o.recoveryMiddleware(wrappedHandler)

	srv := &http.Server{
		Handler: wrappedHandler,
	}

	for _, opt := range o.serverOptions {
		opt(srv)
	}

	return srv
}

func ListenAndServeContext(ctx context.Context, addr string, srv *http.Server) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	go func() {
		if err := srv.Serve(lis); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	return nil
}

func (o *ServerOptions) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o.logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)
		next.ServeHTTP(w, r)
	})
}

func (o *ServerOptions) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				o.panicHandler(w, r, p)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
