package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/bopoh24/bazacars/internal/model"
	"github.com/bopoh24/bazacars/internal/parser"
	"log/slog"
	"math/rand"
	"net/url"
	"strconv"
	"time"
)

const carUrl = "/car-motorbikes-boats-and-parts/cars-trucks-and-vans/"

type Repository interface {
	SaveCars(car []model.Car) error
	Close(ctx context.Context) error
}

type CarParsingService struct {
	targetSite  string
	brands      map[string]string
	repo        Repository
	parsingDate time.Time
	log         *slog.Logger
}

// NewCarParsingService creates a new car parsing service
func NewCarParsingService(targetSite string, repo Repository, log *slog.Logger) *CarParsingService {
	return &CarParsingService{
		targetSite: targetSite,
		repo:       repo,
		log:        log,
	}
}

// LoadCarBrands loads car brands from the target site
func (s *CarParsingService) LoadCarBrands() (err error) {
	s.brands, err = parser.ParseCarBrands(s.targetSite + carUrl)
	return nil
}

// CarBrands returns the list of car brands
func (s *CarParsingService) CarBrands() map[string]string {
	return s.brands
}

// ParseAdsByBrand parses ads by brand
func (s *CarParsingService) ParseAdsByBrand(brand string) error {
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

		s.log.Info("Parsing", "manufacturer", brand, "link", brandPageUrl.String())
		adLinks, err := parser.ParseAdList(brandPageUrl.String())
		if err != nil {
			return err
		}
		pageAds := make([]model.Car, 0, len(adLinks))
		for _, adLink := range adLinks {
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
		if err := s.repo.SaveCars(pageAds); err != nil {
			return err
		}
		s.log.Info("Saved", "brand", brand, "page", i)
	}
	return nil
}

// Close closes the car parsing service
func (s *CarParsingService) Close(ctx context.Context) {
	s.log.Info("closing car parsing service")
	s.repo.Close(ctx)
}
