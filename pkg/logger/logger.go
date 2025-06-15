package logger

import (
	"context"
	"os"
	"production-go-api-template/pkg/contextkeys"

	"github.com/rs/zerolog"
)

type Logger struct {
	zerolog.Logger
}

func New(level zerolog.Level) *Logger {
	zerolog.SetGlobalLevel(level)
	l := zerolog.New(os.Stdout).With().Timestamp().Str("service", "production-go-api-template").Logger()
	return &Logger{l}
}

func (l *Logger) WithRequestID(ctx context.Context) *Logger {
	requestID := contextkeys.GetRequestID(ctx)
	if requestID != "" {
		logger := l.With().Str("request_id", requestID).Logger()
		return &Logger{logger}
	}
	return l
}

func (l *Logger) Tracef(format string, v ...any) {
	l.Trace().Msgf(format, v...)
}

func (l *Logger) Debugf(format string, v ...any) {
	l.Debug().Msgf(format, v...)
}

func (l *Logger) Infof(format string, v ...any) {
	l.Info().Msgf(format, v...)
}

func (l *Logger) Warnf(format string, v ...any) {
	l.Warn().Msgf(format, v...)
}

func (l *Logger) Errorf(format string, v ...any) {
	l.Error().Msgf(format, v...)
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.Fatal().Msgf(format, v...)
}
