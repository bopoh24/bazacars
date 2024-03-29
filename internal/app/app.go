package app

import (
	"context"
	"github.com/bopoh24/bazacars/internal/config"
	"github.com/bopoh24/bazacars/internal/repository"
	"github.com/bopoh24/bazacars/internal/service"
	"github.com/robfig/cron/v3"
	"log/slog"
	"os"
	"time"
)

type App struct {
	conf   *config.Config
	parser *service.CarParsingService
	log    *slog.Logger
}

func New(conf *config.Config, log *slog.Logger) *App {
	repo, err := repository.New(conf.Postgres)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	return &App{
		log:    log,
		conf:   conf,
		parser: service.NewCarParsingService(conf.App.TargetSite, repo, log),
	}
}

func (a *App) Run(ctx context.Context) error {
	a.log.Info("app running")
	c := cron.New()

	// add cron jobs here
	_, err := c.AddFunc("0 11 * * *", func() {
		started := time.Now()
		a.log.Info("Parsing started")
		if err := a.parser.LoadCarBrands(); err != nil {
			a.log.Error("Failed to load car brands", "err", err)
			return
		}
		for brand := range a.parser.CarBrands() {
			if err := a.parser.ParseAdsByBrand(brand); err != nil {
				a.log.Error("Failed to parse ads by brand", "brand", brand, "err", err)
			}
		}
		slog.Info("Parsing finished!", "time", time.Since(started))
	})
	if err != nil {
		return err
	}
	c.Start()
	return nil
}

// Close closes the app
func (a *App) Close(ctx context.Context) {
	a.parser.Close(ctx)
}
