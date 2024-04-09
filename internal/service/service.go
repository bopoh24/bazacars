package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/bopoh24/bazacars/internal/model"
	"github.com/bopoh24/bazacars/internal/parser"
	"github.com/bopoh24/bazacars/internal/repository"
	"log/slog"
	"math/rand"
	"net/url"
	"strconv"
	"time"
)

const carUrl = "/car-motorbikes-boats-and-parts/cars-trucks-and-vans/"

type CarParsingService struct {
	targetSite  string
	brands      map[string]string
	repo        repository.Repository
	parsingDate time.Time
	log         *slog.Logger
}

// NewCarParsingService creates a new car parsing service
func NewCarParsingService(targetSite string, repo repository.Repository, log *slog.Logger) *CarParsingService {
	log = log.With(slog.String("service", "car_parsing"))
	return &CarParsingService{
		targetSite: targetSite,
		repo:       repo,
		log:        log,
	}
}

// LoadCarBrands loads car brands from the target site
func (s *CarParsingService) LoadCarBrands() (err error) {
	s.log.Info("Loading car brands", "url", s.targetSite+carUrl)
	s.brands, err = parser.ParseCarBrands(s.targetSite + carUrl)
	return err
}

// CarBrands returns the list of car brands
func (s *CarParsingService) CarBrands() map[string]string {
	return s.brands
}

// ParseAdsByBrand parses ads by brand
func (s *CarParsingService) ParseAdsByBrand(ctx context.Context, brand string) error {
	brandLink, ok := s.brands[brand]
	if !ok {
		return fmt.Errorf("brand %q not found", brand)
	}
	brandPage, err := url.JoinPath(s.targetSite, brandLink)
	if err != nil {
		return err
	}
	brandUrl, err := url.Parse(brandPage)
	if err != nil {
		return err
	}
	pages, err := parser.TotalPages(brandPage)
	if err != nil {
		return err
	}

	s.log.Info("Total pages", "manufacturer", brand, "pages", pages)

	for i := 1; i <= pages; i++ {
		brandPageUrl := new(url.URL)
		*brandPageUrl = *brandUrl
		query := brandPageUrl.Query()
		query.Add("page", strconv.Itoa(i))
		brandPageUrl.RawQuery = query.Encode()
		if ctx.Err() != nil {
			return ctx.Err()
		}
		s.log.Info("Parsing", "manufacturer", brand, "link", brandPageUrl.String())
		adLinks, err := parser.ParseAdList(brandPageUrl.String())
		if err != nil {
			return err
		}
		pageAds := make([]model.Car, 0, len(adLinks))
		for _, adLink := range adLinks {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			for {
				car, err := parser.ParseCarPage(s.targetSite + adLink)
				if err != nil {
					if errors.Is(err, parser.ErrStatusForbidden) {
						time.Sleep(time.Second + time.Duration(rand.Intn(2500))*time.Millisecond)
						continue
					}
					s.log.Error("Parsing error", "error", err, "link", adLink)
					break
				}
				pageAds = append(pageAds, car)
				break
			}
		}
		if err := s.repo.SaveCars(ctx, pageAds); err != nil {
			return err
		}
		s.log.Info("Saved", "brand", brand, "page", i)
	}
	return nil
}

// NewAds returns new ads
func (s *CarParsingService) NewAds(ctx context.Context) ([]model.Car, error) {
	return s.repo.NewAds(ctx)
}

// AdSent marks the ad as sent
func (s *CarParsingService) AdSent(ctx context.Context, adId string) error {
	return s.repo.AdSent(ctx, adId)
}

// UpdateSent updates the sent field in the database
func (s *CarParsingService) UpdateSent(ctx context.Context) error {
	return s.repo.UpdateSent(ctx)
}

// AdsWithNewPrice returns ads with new price
func (s *CarParsingService) AdsWithNewPrice(ctx context.Context) ([]model.Car, error) {
	return s.repo.AdsWithNewPrice(ctx)
}

// Close closes the car parsing service
func (s *CarParsingService) Close(ctx context.Context) {
	s.log.Info("closing car parsing service")
	s.repo.Close(ctx)
}
