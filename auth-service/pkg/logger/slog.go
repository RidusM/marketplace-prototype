package logger

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	slogotel "github.com/remychantenay/slog-otel"
	slogformatter "github.com/samber/slog-formatter"
	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/pkg/kafka"
)

type Logger struct {
	*slog.Logger
	asyncProducer *kafka.AsyncProducer
}

func New(env string, kafkaBrokers []string) *Logger {
	var log *slog.Logger

	switch env {
	case "local":
		log = slog.New(
			slogformatter.NewFormatterHandler(
				slogformatter.TimezoneConverter(time.UTC),
				slogformatter.TimeFormatter(time.RFC3339, nil),
			)(slogotel.OtelHandler{
				Next: slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
					Level:     slog.LevelDebug,
					AddSource: false,
				}),
				NoBaggage:     false,
				NoTraceEvents: false,
			}))

	case "dev":
		log = slog.New(
			slogformatter.NewFormatterHandler(
				slogformatter.TimezoneConverter(time.UTC),
				slogformatter.TimeFormatter(time.RFC3339, nil),
			)(slogotel.OtelHandler{
				Next: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
					Level:     slog.LevelDebug,
					AddSource: false,
				}),
				NoBaggage:     false,
				NoTraceEvents: false,
			}))

	case "prod":
		log = slog.New(
			slogformatter.NewFormatterHandler(
				slogformatter.TimezoneConverter(time.UTC),
				slogformatter.TimeFormatter(time.RFC3339, nil),
			)(slogotel.OtelHandler{
				Next: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
					Level:     slog.LevelInfo,
					AddSource: false,
				}),
				NoBaggage:     false,
				NoTraceEvents: false,
			}))
	}

	if kafkaBrokers != nil || len(kafkaBrokers) > 0 {
		asyncProducer, err := kafka.NewAsyncProducer(kafkaBrokers, "application.logs")
		if err != nil {
			log.Warn("logger.New: failed while create new async producer", Err(err))
		}

		return &Logger{
			Logger:        log,
			asyncProducer: asyncProducer,
		}
	}

	return &Logger{
		Logger: log,
	}
}

func (l *Logger) LogToKafka(level slog.Level, message string) {
	if l.asyncProducer != nil {
		l.asyncProducer.SendMessage(level.String(), message)
	}
}

func (l *Logger) logWithSource(level slog.Level, msg string, args ...any) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}

	source := fmt.Sprintf("%s:%d", file, line)
	args = append(args, slog.String("source", source))

	switch level {
	case slog.LevelInfo:
		l.Logger.Info(msg, args...)
	case slog.LevelError:
		l.Logger.Error(msg, args...)
	case slog.LevelWarn:
		l.Logger.Warn(msg, args...)
	case slog.LevelDebug:
		l.Logger.Debug(msg, args...)
	}

	l.LogToKafka(level, msg)
}

func (l *Logger) Info(msg string, args ...any) {
	l.logWithSource(slog.LevelInfo, msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.logWithSource(slog.LevelError, msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logWithSource(slog.LevelWarn, msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.logWithSource(slog.LevelDebug, msg, args...)
}

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "Error",
		Value: slog.StringValue(err.Error()),
	}
}
