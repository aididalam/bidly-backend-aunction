package main

import (
	"auction/auction/internal/config"
	"auction/auction/internal/handler"
	"auction/auction/internal/middleware"
	"auction/auction/internal/repository"
	"auction/auction/internal/service"
	"auction/auction/internal/token"
	"auction/auction/internal/upload"
	"auction/auction/internal/worker"
	"auction/auction/migrations"
	"context"
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	db, err := sql.Open("mysql", cfg.DatabaseURL)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(3 * time.Minute)
	startup, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err = db.PingContext(startup); err != nil {
		logger.Error("connect database", "error", err)
		os.Exit(1)
	}
	if err = migrations.Up(startup, db); err != nil {
		logger.Error("run migrations", "error", err)
		os.Exit(1)
	}
	uploads, err := upload.New(startup, cfg.AWSRegion, cfg.S3Bucket, cfg.S3PublicBaseURL, cfg.S3Endpoint, cfg.S3UsePathStyle)
	if err != nil {
		logger.Error("initialize S3", "error", err)
		os.Exit(1)
	}
	svc := service.New(repository.NewMySQL(db), uploads)
	tokens := token.New(cfg.JWTSecret)
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go worker.RunExpiry(rootCtx, svc, cfg.ExpireInterval, logger)
	server := &http.Server{Addr: cfg.Address(), Handler: handler.New(svc, middleware.New(tokens)), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	errs := make(chan error, 1)
	go func() {
		logger.Info("auction service started", "address", server.Addr)
		errs <- server.ListenAndServe()
	}()
	select {
	case <-rootCtx.Done():
	case err = <-errs:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped", "error", err)
			os.Exit(1)
		}
	}
	ctx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()
	if err = server.Shutdown(ctx); err != nil {
		logger.Error("shutdown", "error", err)
		os.Exit(1)
	}
}
