package database

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"recipe-backend/internal/pkg/requestid"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// slogGormLogger は GORM の logger.Interface を slog で実装する。
// 通常クエリは Debug、遅いクエリは Warn、エラーは Error で出力し、
// context に request_id があれば付与してリクエストと相関できるようにする。
type slogGormLogger struct {
	logger        *slog.Logger
	slowThreshold time.Duration
}

func NewGormLogger(logger *slog.Logger) gormlogger.Interface {
	return &slogGormLogger{
		logger:        logger,
		slowThreshold: 200 * time.Millisecond,
	}
}

func (l *slogGormLogger) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *slogGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.logger.InfoContext(ctx, msg, withRequestID(ctx, slog.Any("data", data))...)
}

func (l *slogGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logger.WarnContext(ctx, msg, withRequestID(ctx, slog.Any("data", data))...)
}

func (l *slogGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logger.ErrorContext(ctx, msg, withRequestID(ctx, slog.Any("data", data))...)
}

func (l *slogGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	attrs := withRequestID(ctx,
		slog.String("sql", sql),
		slog.Int64("rows", rows),
		slog.Duration("elapsed", elapsed),
	)

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		l.logger.ErrorContext(ctx, "gorm query", append(attrs, slog.String("error", err.Error()))...)
	case elapsed > l.slowThreshold:
		l.logger.WarnContext(ctx, "gorm slow query", attrs...)
	default:
		l.logger.DebugContext(ctx, "gorm query", attrs...)
	}
}

// withRequestID は context に request_id があれば先頭に付与した属性スライスを返す。
func withRequestID(ctx context.Context, attrs ...any) []any {
	if rid := requestid.FromContext(ctx); rid != "" {
		return append([]any{slog.String("request_id", rid)}, attrs...)
	}
	return attrs
}
