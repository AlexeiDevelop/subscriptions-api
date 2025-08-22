// @title           Subscriptions API
// @version         1.0
// @description     REST-сервис для агрегации онлайн-подписок пользователей.
// @BasePath        /
// @schemes         http

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AlexeiDevelop/subscriptions-api/internal/config"
	"github.com/AlexeiDevelop/subscriptions-api/internal/handler"
	"github.com/AlexeiDevelop/subscriptions-api/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/AlexeiDevelop/subscriptions-api/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	lg := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		lg.Error("config", slog.Any("err", err))
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := storage.NewPostgresPool(ctx, cfg.DB.DSN())
	if err != nil {
		lg.Error("db connect", slog.Any("err", err))
		os.Exit(1)
	}
	defer pool.Close()

	repo := storage.NewRepository(pool)
	h := handler.New(repo, lg)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK); w.Write([]byte("ok")) })
	h.RegisterRoutes(r)
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		lg.Info("server_start", slog.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lg.Error("server", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	// Graceful_shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctxShutdown)
	lg.Info("server_stopped")
}
