package logger

import (
	"context"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/logger/config"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/logger/core"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/context"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type DualLogger struct {
	logger *zap.Logger
}

func NewDualLogger(cfg *config.LogConfig) (*DualLogger, error) {
	builder, err := core.NewCoreBuilder(cfg)
	if err != nil {
		return nil, err
	}

	newLogger, err := builder.DualLogger()
	if err != nil {
		return nil, err
	}

	return &DualLogger{
		logger: newLogger,
	}, nil
}

func (d *DualLogger) log(ctx context.Context, level zapcore.Level, msg string, opts ...logger.LogOption) {
	additional := make(map[string]any)
	for _, opt := range opts {
		opt(additional)
	}

	correlationID, ok := ctx.Value(pkg.CorrelationID).(string)
	if !ok {
		correlationID = ""
	}

	fields := make([]zap.Field, 0, len(additional)+1)
	for k, v := range additional {
		fields = append(fields, zap.Any(k, v))
	}
	if correlationID != "" {
		fields = append(fields, zap.String(pkg.CorrelationID.String(), correlationID))
	}

	d.logger.Log(level, msg, fields...)
}

func (d *DualLogger) Debug(ctx context.Context, msg string, opts ...logger.LogOption) {
	d.log(ctx, zapcore.DebugLevel, msg, opts...)
}

func (d *DualLogger) Info(ctx context.Context, msg string, opts ...logger.LogOption) {
	d.log(ctx, zapcore.InfoLevel, msg, opts...)
}

func (d *DualLogger) Warn(ctx context.Context, msg string, opts ...logger.LogOption) {
	d.log(ctx, zapcore.WarnLevel, msg, opts...)
}

func (d *DualLogger) Error(ctx context.Context, msg string, opts ...logger.LogOption) {
	d.log(ctx, zapcore.ErrorLevel, msg, opts...)
}

func (d *DualLogger) DPanic(ctx context.Context, msg string, opts ...logger.LogOption) {
	d.log(ctx, zapcore.DPanicLevel, msg, opts...)
}

func (d *DualLogger) Panic(ctx context.Context, msg string, opts ...logger.LogOption) {
	d.log(ctx, zapcore.PanicLevel, msg, opts...)
}

func (d *DualLogger) Fatal(ctx context.Context, msg string, opts ...logger.LogOption) {
	d.log(ctx, zapcore.FatalLevel, msg, opts...)
}
