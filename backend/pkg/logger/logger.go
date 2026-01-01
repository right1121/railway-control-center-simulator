package logger

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"

	"github.com/right1121/railway-control-center-simulator/pkg/apperr"
)

// ログレベルの定数
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

var defaultLogger *Logger

type Logger struct {
	logger *slog.Logger
}

type ctxKey = struct{}

// New は新しいロガーインスタンスを作成します
func New(opts ...Option) *Logger {
	options := &options{
		level:  LevelInfo,
		format: "json",
	}

	for _, opt := range opts {
		opt(options)
	}

	var handler slog.Handler
	if options.format == "text" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: options.level,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: options.level,
		})
	}

	return &Logger{
		logger: slog.New(handler),
	}
}

func SetDefault(logger *Logger) {
	defaultLogger = logger
}

func GetDefault() *Logger {
	if defaultLogger == nil {
		SetDefault(New())
	}
	return defaultLogger
}

// WithContext はコンテキストに関連付けられたロガーを返します
func (l *Logger) WithContext(ctx context.Context) *Logger {
	traceID := GetTraceID(ctx)
	if traceID == "" {
		// トレースIDが存在しない場合は新規作成
		ctx = WithTraceID(ctx)
		traceID = GetTraceID(ctx)
	}

	return &Logger{
		logger: l.logger.With(
			slog.String("trace_id", traceID),
		),
	}
}

// WithRequestInfo はリクエスト情報をコンテキストに追加したロガーを返します
func (l *Logger) WithRequestInfo(ctx context.Context, info *RequestInfo) *Logger {
	// リクエスト情報をコンテキストに保存
	_ = context.WithValue(ctx, ctxKey{}, info)

	// ロガーにリクエスト情報を追加
	return &Logger{
		logger: l.logger.With(
			slog.Group("request",
				slog.String("method", info.Method),
				slog.String("path", info.Path),
				slog.String("query", info.Query),
				slog.String("remote_addr", info.RemoteAddr),
				slog.String("user_agent", info.UserAgent),
				slog.String("protocol", info.Protocol),
			),
		),
	}
}

// GetRequestInfo はコンテキストからリクエスト情報を取得します
func GetRequestInfo(ctx context.Context) *RequestInfo {
	if info, ok := ctx.Value(ctxKey{}).(*RequestInfo); ok {
		return info
	}
	return nil
}

// オプション設定用の型と関数
type options struct {
	level  slog.Level
	format string
}

type Option func(*options)

// WithLevel はログレベルを設定するオプションを返します
func WithLevel(level slog.Level) Option {
	return func(o *options) {
		o.level = level
	}
}

// WithFormat はログフォーマットを設定するオプションを返します
// format: "json" or "text"
func WithFormat(format string) Option {
	return func(o *options) {
		o.format = format
	}
}

func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	argsWithStack := append([]any{slog.Any("stack", stack())}, args...)
	l.logger.Error(msg, argsWithStack...)
}

// WithError は AppError を構造化して出力します
func (l *Logger) WithError(err error) {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		args := []any{
			slog.Any("stack", stack()),
			unwrapChainAttr(appErr),
		}
		l.logger.Error(appErr.Error(), args...)
	} else {
		// fallback: 通常のエラーとしてログ出力
		args := []any{
			slog.String("error", err.Error()),
		}
		l.Error("unexpected error", args...)
	}
}

func unwrapChainAttr(err error) slog.Attr {
	var chain []any

	for err != nil {
		var appErr *apperr.AppError
		if errors.As(err, &appErr) {
			chain = append(chain, map[string]any{
				"message":   appErr.Message,
				"operation": appErr.Operation,
				"code":      appErr.Code,
			})
		} else {
			chain = append(chain, map[string]any{
				"message": err.Error(),
			})
		}
		err = errors.Unwrap(err)
	}

	return slog.Any("errors", chain)
}

// stack はスタックトレースを取得します
func stack() []string {
	// スタックトレース情報を取得
	stackTrace := make([]string, 0)
	for i := 1; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		stackTrace = append(stackTrace, fmt.Sprintf("%s:%d", file, line))
	}

	return stackTrace
}