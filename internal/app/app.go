package app

import (
	"context"
	"fmt"
	"github.com/bopoh24/bazacars/internal/bot"
	"github.com/bopoh24/bazacars/internal/config"
	"github.com/bopoh24/bazacars/internal/model"
	"github.com/bopoh24/bazacars/internal/repository/postgres"
	"github.com/bopoh24/bazacars/internal/service"
	"github.com/robfig/cron/v3"
	"log/slog"
	"os"
	"time"
)

const (
	EmojiAlert     = "🚨"
	EmojiNew       = "🆕"
	EmojiCar       = "🚗"
	EmojiEuro      = "💶"
	EmojiLocation  = "📍"
	EmojiDate      = "📅"
	EmojiArrow     = "➡️"
	EmojiWarning   = "⚠️"
	EmojiChartDown = "📉"
	EmojiChartUp   = "📈"
)

type App struct {
	conf   *config.Config
	parser *service.CarParsingService
	bot    *bot.Bot
	log    *slog.Logger
}

func New(conf *config.Config, log *slog.Logger) *App {
	repo, err := postgres.NewRepository(conf.Postgres)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	tgBot, err := bot.New(conf.Token.TelegramBotToken, repo, log)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	return &App{
		log:    log,
		conf:   conf,
		bot:    tgBot,
		parser: service.NewCarParsingService(conf.App.TargetSite, repo, log),
	}
}

func (a *App) Run(ctx context.Context) error {
	a.log.Info("app running")
	c := cron.New()

	go a.bot.Run(ctx)

	// add cron jobs here
	_, err := c.AddFunc("5 12 * * *", func() {
		started := time.Now()
		a.log.Info("Parsing started")
		if err := a.parser.LoadCarBrands(); err != nil {
			a.log.Error("Failed to load car brands", "err", err)
			return
		}
		for brand := range a.parser.CarBrands() {
			if err := a.parser.ParseAdsByBrand(ctx, brand); err != nil {
				a.log.Error("Failed to parse ads by brand", "brand", brand, "err", err)
			}
		}
		a.log.Info("Parsing finished!", "time", time.Since(started))
		a.log.Info("Sending new ads to subscribers")
		ads, err := a.parser.NewAds(ctx)
		if err != nil {
			a.log.Error("Failed to get new ads", "err", err)
			return
		}
		if len(ads) == 0 {
			a.log.Info("No new ads")
			return
		}
		for _, ad := range ads {
			err = a.bot.SendMessageToSubscribers(ctx, newCarMessage(ad))
			if err != nil {
				a.log.Error("Failed to send ad", "err", err)
			}
			err = a.parser.AdSent(ctx, ad.AdID)
			if err != nil {
				a.log.Error("Failed to mark ad as sent", "err", err)
			}
		}
		a.log.Info("New ads sent")
		a.log.Info("Sending ads with new price to subscribers")
		cars, err := a.parser.AdsWithNewPrice(ctx)
		if err != nil {
			a.log.Error("Failed to get ads with new price", "err", err)
			return
		}
		if len(cars) == 0 {
			a.log.Info("No ads with new price")
			return
		}
		for _, car := range cars {
			err = a.bot.SendMessageToSubscribers(ctx, priceChangedMessage(car))
			if err != nil {
				a.log.Error("Failed to send ad", "err", err)
			}
			err = a.parser.AdSent(ctx, car.AdID)
			if err != nil {
				a.log.Error("Failed to mark ad as sent", "err", err)
			}
		}
		a.log.Info("Ads with new price sent")
	})
	if err != nil {
		return err
	}
	c.Start()
	return nil
}

func newCarMessage(c model.Car) string {
	return fmt.Sprintf("%s <strong>%s %s</strong> (%d)\n\n"+
		"%s <strong>%d€</strong>\n\n"+
		"%s %dkm (%s)\n\n<i>%s %s</i>\n%s\n%s",
		EmojiNew, c.Manufacturer, c.Model, c.Year, EmojiEuro, c.Price, EmojiCar,
		c.Mileage, c.Fuel, EmojiLocation, c.Address, c.Posted.Format("02.01.2006 15:04"), c.Link)
}

func priceChangedMessage(c model.Car) string {

	arrEmoji := EmojiChartDown
	if c.Price > c.OldPrice {
		arrEmoji = EmojiChartUp
	}
	return fmt.Sprintf(
		"%s <strong>%s %s</strong>  (%d)\n\n"+
			"%s <s>%d€</s> %s <strong>%d€</strong>\n\n"+
			"%s %dkm (%s)\n\n<i>%s %s</i>\n%s\n%s",
		arrEmoji, c.Manufacturer, c.Model, c.Year, EmojiEuro, c.OldPrice, EmojiArrow, c.Price, EmojiCar,
		c.Mileage, c.Fuel, EmojiLocation, c.Address, c.Posted.Format("02.01.2006 15:04"), c.Link)
}

// Close closes the app
func (a *App) Close(ctx context.Context) {
	a.parser.Close(ctx)
}
