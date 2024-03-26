package repository

import (
	"context"
	"fmt"
	"github.com/bopoh24/bazacars/internal/config"
	"github.com/bopoh24/bazacars/internal/model"
	"github.com/bopoh24/bazacars/pkg/sql/builder"
)

type Repository struct {
	psql *builder.Postgres
}

// New returns a new Repository struct
func New(dbConf config.Postgres) (*Repository, error) {
	psqlConn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbConf.Host, dbConf.Port, dbConf.User, dbConf.Pass, dbConf.Database)
	psql, err := builder.NewPostgresBuilder(psqlConn)
	if err != nil {
		return nil, err
	}
	return &Repository{psql: psql}, nil

}

func (r *Repository) SaveCars(cars []model.Car) error {
	q := r.psql.Builder().Insert("cars").Columns("manufacturer", "model", "year", "mileage", "engine",
		"fuel", "drive", "automatic", "power", "color", "price", "description", "ad_id", "address", "link", "posted")
	for _, car := range cars {
		q = q.Values(car.Manufacturer, car.Model, car.Year, car.Mileage, car.EngineSize, car.Fuel, car.Drive,
			car.AutomaticGearbox, car.Power, car.Color, car.Price, car.Description, car.AdID,
			car.Address, car.Link, car.Posted)
	}
	q = q.Suffix("ON CONFLICT (ad_id, parsed) DO NOTHING")
	_, err := q.Exec()
	if err != nil {
		return err
	}
	return nil
}

// Close closes the repository
func (r *Repository) Close(ctx context.Context) error {
	return r.psql.Close()
}
