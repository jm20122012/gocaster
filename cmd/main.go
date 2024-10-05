package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/jm20122012/gocaster/internal/config"
)

var (
	logger      *slog.Logger
	logLevelVar slog.LevelVar
)

func main() {
	createLogger("debug")

	logger.Debug("Loading config")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger.Debug("Config loaded", "config", cfg)

	createLogger(cfg.DebugLevel)

	logger.Info("Starting GoCaster server")

}

func createLogger(level string) {

	switch level {
	case "debug":
		logLevelVar.Set(slog.LevelDebug)
	case "info":
		logLevelVar.Set(slog.LevelInfo)
	case "warn":
		logLevelVar.Set(slog.LevelWarn)
	case "error":
		logLevelVar.Set(slog.LevelError)
	default:
		logLevelVar.Set(slog.LevelInfo)
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: &logLevelVar,
		// AddSource: true,
	})

	logger = slog.New(handler)

}
