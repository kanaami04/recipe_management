package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "time/tzdata" // Asia/Tokyo を埋め込み（Windows でも LoadLocation を可能にする）

	"recipe-backend/internal/app"
	"recipe-backend/internal/config"
	"recipe-backend/internal/database"
	"recipe-backend/internal/logger"
)

// shutdownTimeout はグレースフルシャットダウン時に処理中リクエストを待つ最大時間。
const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		slog.Error("startup failed", "error", err)
		os.Exit(1) // 終了は1か所だけ
	}
}

func run() error {
	// 例: go run main.go -env .env.local
	envFile := flag.String("env", ".env", "path to env file")
	flag.Parse()

	cfg := config.Load(*envFile)

	log := logger.New(cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(log)

	// インフラ（DB）の用意は main が担い、各層の配線は app（合成ルート）に委ねる。
	db, err := database.Connect(cfg.DatabaseURL, log)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}
	defer sqlDB.Close() // 終了時に DB プールを閉じる（サーバ停止の後に走る）

	if err := database.Migrate(db); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	e := app.New(cfg, db, log)

	// SIGINT / SIGTERM を待ち受ける context。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// サーバは別 goroutine で起動し、致命的な起動エラーだけを通知する。
	serverErr := make(chan error, 1)
	go func() {
		log.Info("server starting", "port", cfg.Port)
		if err := e.Start(":" + cfg.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// 「起動エラー」か「シャットダウンシグナル」のどちらかを待つ。
	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		log.Info("shutdown signal received, draining connections...")
	}

	// グレースフルシャットダウン：処理中リクエストを shutdownTimeout まで捌く。
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	log.Info("server stopped gracefully")
	return nil
}
