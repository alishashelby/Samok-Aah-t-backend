package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/logger/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const fileMode = 0644

type CoreBuilder struct {
	level zapcore.Level
	cfg   *config.LogConfig
}

func NewCoreBuilder(cfg *config.LogConfig) (*CoreBuilder, error) {
	level := zapcore.InfoLevel
	if err := level.Set(cfg.Logger.Level); err != nil {
		return nil, fmt.Errorf("failed to set log level: %w", err)
	}

	return &CoreBuilder{
		level: level,
		cfg:   cfg,
	}, nil
}

func (c *CoreBuilder) ConsoleCore() zapcore.Core {
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)

	consoleLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= c.level
	})

	return zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), consoleLevel)
}

func (c *CoreBuilder) JSONCore() (zapcore.Core, error) {
	if c.cfg.Logger.LogsDir == "" {
		return nil, errors.New("logs dir path is empty")
	}
	if c.cfg.Logger.LogsFile == "" {
		return nil, errors.New("logs file name is empty")
	}

	jsonCfg := zap.NewProductionEncoderConfig()
	jsonCfg.TimeKey = "@timestamp"
	jsonCfg.LevelKey = "log.level"
	jsonCfg.MessageKey = "message"
	jsonCfg.CallerKey = "caller"
	jsonCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	jsonCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	jsonEncoder := zapcore.NewJSONEncoder(jsonCfg)

	if err := os.MkdirAll(c.cfg.Logger.LogsDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("create logs dir err: %w", err)
	}

	logFile := filepath.Join(c.cfg.Logger.LogsDir, c.cfg.Logger.LogsFile)
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, fileMode)
	if err != nil {
		return nil, fmt.Errorf("create log file failed: %w", err)
	}

	jsonLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl != zapcore.DebugLevel
	})

	return zapcore.NewCore(jsonEncoder, zapcore.AddSync(file), jsonLevel), nil
}

func (c *CoreBuilder) DualLogger() (*zap.Logger, error) {
	consoleCore := c.ConsoleCore()

	jsonCore, err := c.JSONCore()
	if err != nil {
		return nil, err
	}

	core := zapcore.NewTee(consoleCore, jsonCore)

	return zap.New(core, zap.AddCaller()), nil
}
