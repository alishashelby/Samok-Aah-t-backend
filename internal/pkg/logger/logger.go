package logger

import "context"

type LogOption func(map[string]any)

type Logger interface {
	Debug(ctx context.Context, msg string, opts ...LogOption)
	Info(ctx context.Context, msg string, opts ...LogOption)
	Warn(ctx context.Context, msg string, opts ...LogOption)
	Error(ctx context.Context, msg string, opts ...LogOption)
	DPanic(ctx context.Context, msg string, opts ...LogOption)
	Panic(ctx context.Context, msg string, opts ...LogOption)
	Fatal(ctx context.Context, msg string, opts ...LogOption)
}
