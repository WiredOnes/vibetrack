package telemetry

import (
	"context"
	"os"

	"github.com/leshless/golibrary/graceful"
	"github.com/leshless/pet/cub/internal/config"
	"github.com/leshless/pet/cub/internal/environment"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/lumberjack.v2"
)

type logger struct {
	zapLogger *zap.Logger
}

var _ Logger = (*logger)(nil)

func (l *logger) Sync() error {
	return l.zapLogger.Sync()
}

func (l *logger) Debug(ctx context.Context, message string, labels ...label) {
	labels = allLabels(ctx, labels)
	l.zapLogger.Debug(message, labelsToZap(labels)...)
}

func (l *logger) Info(ctx context.Context, message string, labels ...label) {
	labels = allLabels(ctx, labels)
	l.zapLogger.Info(message, labelsToZap(labels)...)
}

func (l *logger) Warn(ctx context.Context, message string, labels ...label) {
	labels = allLabels(ctx, labels)
	l.zapLogger.Warn(message, labelsToZap(labels)...)
}

func (l *logger) Error(ctx context.Context, message string, labels ...label) {
	labels = allLabels(ctx, labels)
	l.zapLogger.Error(message, labelsToZap(labels)...)
}

func (l *logger) With(labels ...label) Logger {
	return &logger{l.zapLogger.With(labelsToZap(labels)...)}
}

func InitLogger(
	configHolder config.Holder,
	environmentHolder environment.Holder,
	gracefulRegistrator graceful.Registrator,
) (*logger, error) {
	config := configHolder.Config().Logger
	serviceConfig := configHolder.Config().Service
	environment := environmentHolder.Environment()

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        config.TimeKey,
		LevelKey:       config.LevelKey,
		NameKey:        config.NameKey,
		CallerKey:      config.CallerKey,
		MessageKey:     config.MessageKey,
		StacktraceKey:  config.StacktraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if config.Developement {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	var writer zapcore.WriteSyncer
	if config.StdoutOnly {
		writer = zapcore.AddSync(os.Stdout)
	} else {
		writer = zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.OutputFile,
			MaxSize:    config.MaxFileSizeMB,
			MaxBackups: config.MaxFilesAmount,
			MaxAge:     config.MaxFileAgeDays,
			Compress:   config.Compression,
		})
	}

	var level zapcore.Level
	if config.Developement {
		level = zapcore.DebugLevel
	} else {
		level = zapcore.InfoLevel
	}

	core := zapcore.NewCore(
		encoder,
		writer,
		level,
	)

	baseLabels := []label{
		Service(serviceConfig.Name),
		Environment(serviceConfig.Environment),
		Host(environment.HostName),
	}

	zapLogger := zap.New(core, zap.Fields(labelsToZap(baseLabels)...))

	logger := newLogger(zapLogger)

	logger.Info(context.Background(), "logger successfully initialized")

	gracefulRegistrator.Register(func(_ context.Context) error {
		return logger.Sync()
	})

	return logger, nil
}
