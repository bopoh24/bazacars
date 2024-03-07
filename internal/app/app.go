package app

import (
	"context"
	"github.com/bopoh24/bazacars/internal/config"
	"log/slog"
)

type App struct {
	conf *config.Config
	log  *slog.Logger
}

func New(conf *config.Config, log *slog.Logger) *App {
	return &App{
		log:  log,
		conf: conf,
	}
}

func (a *App) Run(ctx context.Context) error {
	a.log.Info("app running")
	return nil
}

func (a *App) Close(ctx context.Context) {
	a.log.Info("app closed")
}
