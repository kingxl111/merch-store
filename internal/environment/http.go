package http

import (
	"context"
	merchstoreapi "github.com/kingxl111/merch-store/pkg/api/merch-store"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/go-faster/errors"
	"github.com/kingxl111/merch-store/internal/users/service"
)

const UsernameContextKey = "username"

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

func (o *ServerOptions) NewServer(handler http.Handler, addr string) *http.Server {
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
	wrappedHandler = o.authMiddleware(wrappedHandler)

	srv := &http.Server{
		Handler: wrappedHandler,
		Addr:    addr,
	}

	for _, opt := range o.serverOptions {
		opt(srv)
	}

	return srv
}

func ListenAndServeContext(ctx context.Context, srv *http.Server) error {
	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			slog.Error("HTTP server shutdown error", "error", err)
		}
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

func (o *ServerOptions) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.String(), "/api/auth") {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		username, err := service.ParseToken(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, UsernameContextKey, username)
		o.logger.Info("user: " + ctx.Value(UsernameContextKey).(string))
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func NewRouter(handler merchstoreapi.ServerInterface) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/auth", handler.PostApiAuth)
	mux.HandleFunc("/api/buy/", func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем параметр item из пути
		item := r.URL.Path[len("/api/buy/"):]
		handler.GetApiBuyItem(w, r, item)
	})
	mux.HandleFunc("/api/info", handler.GetApiInfo)
	mux.HandleFunc("/api/sendCoin", handler.PostApiSendCoin)

	return mux
}
