package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/keycloak-broker/pkg/config"
)

var level = new(slog.LevelVar)

func Init() {
	cfg := config.Get()
	level.Set(parseLevel(cfg.LogLevel))
	replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
		if !cfg.LogTimestamp && a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		if cfg.LogTimestamp && a.Key == slog.TimeKey {
			a.Value = slog.StringValue(a.Value.Time().Format(time.RFC3339Nano))
		}
		return a
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level, ReplaceAttr: replaceAttr})))
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "fatal":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Debug(format string, v ...interface{}) { slog.Debug(fmt.Sprintf(format, v...)) }
func Info(format string, v ...interface{})  { slog.Info(fmt.Sprintf(format, v...)) }
func Warn(format string, v ...interface{})  { slog.Warn(fmt.Sprintf(format, v...)) }
func Error(format string, v ...interface{}) { slog.Error(fmt.Sprintf(format, v...)) }

func Fatal(format string, v ...interface{}) {
	slog.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}
