package logger

import (
	"log/slog"
	"os"

	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/config"
)

// InitLogger initializes the global structured logger based on the application config.
func InitLogger(cfg config.Config) {
	var logHandler slog.Handler

	level := new(slog.LevelVar)
	switch cfg.LogLevel {
	case "DEBUG":
		level.Set(slog.LevelDebug)
	case "WARN":
		level.Set(slog.LevelWarn)
	case "ERROR":
		level.Set(slog.LevelError)
	default:
		level.Set(slog.LevelInfo)
	}

	handlerOpts := &slog.HandlerOptions{
		Level: level,
	}

	if cfg.LogFormat == "json" {
		logHandler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, handlerOpts)
	}

	logger := slog.New(logHandler)
	slog.SetDefault(logger)
}
