package main

import (
	"context"
	"errors"
	"github.com/bopoh24/bazacars/internal/app"
	"github.com/bopoh24/bazacars/internal/config"
	"github.com/bopoh24/bazacars/pkg/logger"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	// init config
	cfg, err := config.New()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	// init logger
	log := logger.New(cfg.App.LogLevel)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// init app
	a := app.New(cfg, log)
	if err = a.Run(ctx); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			slog.Info("server closed")
			return
		}
		log.Error(err.Error())
		os.Exit(1)
	}

	// graceful shutdown
	<-ctx.Done()
	// closing context
	closeCtx, closeCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer closeCancel()
	a.Close(closeCtx)
}
