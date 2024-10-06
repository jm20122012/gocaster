package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jm20122012/gocaster/internal/config"
	"github.com/jm20122012/gocaster/internal/server"
)

var (
	logger      *slog.Logger
	logLevelVar slog.LevelVar
)

func main() {
	createLogger("debug")

	logger.Debug("Loading config")

	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger.Debug("Config loaded", "config", conf)

	createLogger(conf.DebugLevel)

	logger.Info("Starting GoCaster server")

	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("Ctrl+C pressed, cancelling context...")
		cancel()
	}()

	s, err := server.NewServer(
		ctx,
		cancel,
		logger,
		conf,
	)
	if err != nil {
		logger.Error("error creating gocaster server", "error", err)
		cancel()
	}

	s.Start()

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
