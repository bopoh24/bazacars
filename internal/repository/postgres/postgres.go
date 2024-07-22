package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/bopoh24/bazacars/internal/config"
	"github.com/bopoh24/bazacars/internal/model"
	"github.com/bopoh24/bazacars/internal/repository"
	"github.com/bopoh24/bazacars/pkg/sql/builder"
	"time"
)

var favouriteBrands = []string{
	"BMW", "Mercedes-Benz", "Mazda", "Toyota", "Nissan",
	"Audi", "Volkswagen", "Ford", "Honda", "Lexus",
	"Jeep", "Volvo", "Infiniti", "Acura", "Land Rover", "Jaguar", "Mini",
}

var excludeModels = []string{"Fit", "2", "Yaris", "Aqua", "Sienta", "Polo", "CX-3", "Voxy", "Porte", "Yaris Cross", "C-HR", "Vezel"}

const (
	maxPrice      = 29000
	minYear       = 2020
	maxMileage    = 50000
	minEngineSize = 1.5
)

type Repository struct {
	psql *builder.Postgres
}

// NewRepository returns a new Repository struct
func NewRepository(dbConf config.Postgres) (*Repository, error) {
	psqlConn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbConf.Host, dbConf.Port, dbConf.User, dbConf.Pass, dbConf.Database)
	psql, err := builder.NewPostgresBuilder(psqlConn)
	if err != nil {
		return nil, err
	}
	return &Repository{psql: psql}, nil

}

// SaveCars saves cars to the database
func (r *Repository) SaveCars(ctx context.Context, cars []model.Car) error {
	q := r.psql.Builder().Insert("cars").Columns("manufacturer", "model", "year", "mileage", "engine",
		"fuel", "drive", "automatic", "power", "color", "price", "description", "ad_id", "address", "link", "posted")
	for _, car := range cars {
		q = q.Values(car.Manufacturer, car.Model, car.Year, car.Mileage, car.EngineSize, car.Fuel, car.Drive,
			car.AutomaticGearbox, car.Power, car.Color, car.Price, car.Description, car.AdID,
			car.Address, car.Link, car.Posted)
	}
	q = q.Suffix("ON CONFLICT (ad_id, parsed) DO NOTHING")
	_, err := q.ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Users returns all users
func (r *Repository) Users(ctx context.Context) ([]model.User, error) {
	q := r.psql.Builder().Select("*").From("users")
	rows, err := q.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []model.User
	for rows.Next() {
		var user model.User
		if err = rows.Scan(&user.ChatID, &user.Username, &user.FirstName, &user.LastName,
			&user.Approved, &user.Admin, &user.UpdatedAt, &user.CreatedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, repository.ErrNotFound
			}
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *Repository) User(ctx context.Context, chatID int64) (model.User, error) {
	q := r.psql.Builder().Select("*").From("users").Where(sq.Eq{"chat_id": chatID})
	row := q.QueryRowContext(ctx)
	var user model.User
	if err := row.Scan(&user.ChatID, &user.Username, &user.FirstName, &user.LastName, &user.Approved,
		&user.Admin, &user.UpdatedAt, &user.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, repository.ErrNotFound
		}
		return model.User{}, err
	}
	return user, nil
}

func (r *Repository) Admins(ctx context.Context) ([]model.User, error) {
	q := r.psql.Builder().Select("*").From("users").Where(sq.Eq{"admin": true})
	rows, err := q.QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	defer rows.Close()
	var users []model.User
	for rows.Next() {
		var user model.User
		err = rows.Scan(&user.ChatID, &user.Username, &user.FirstName, &user.LastName,
			&user.Approved, &user.Admin, &user.UpdatedAt, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *Repository) UserAdd(ctx context.Context, user model.User) error {
	q := r.psql.Builder().Insert("users").Columns("chat_id", "username", "first_name", "last_name",
		"approved", "admin", "updated_at", "created_at")
	q = q.Values(user.ChatID, user.Username, user.FirstName, user.LastName, user.Approved, user.Admin,
		user.UpdatedAt, user.CreatedAt)
	_, err := q.ExecContext(ctx)
	return err
}

func (r *Repository) UserSave(ctx context.Context, user model.User) error {
	q := r.psql.Builder().Update("users").
		Set("username", user.Username).
		Set("first_name", user.FirstName).
		Set("last_name", user.LastName).
		Set("approved", user.Approved).
		Set("admin", user.Admin).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"chat_id": user.ChatID})
	_, err := q.ExecContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.ErrNotFound
		}
	}
	return err
}

func (r *Repository) NewAds(ctx context.Context) ([]model.Car, error) {
	q := r.psql.Builder().Select("manufacturer", "model", "year", "mileage", "engine", "fuel", "drive", "automatic",
		"power", "color", "price", "description", "ad_id", "address", "link", "posted").Distinct().
		From("cars").
		Where(
			sq.And{
				sq.Eq{"manufacturer": favouriteBrands},
				sq.NotEq{"model": excludeModels},
				sq.LtOrEq{"price": maxPrice},
				sq.LtOrEq{"mileage": maxMileage},
				sq.GtOrEq{"year": minYear},
				sq.GtOrEq{"engine": minEngineSize},
				sq.Eq{"automatic": true},
				sq.GtOrEq{"posted": time.Now().AddDate(0, 0, -1).Format("2006-01-02")},
				sq.GtOrEq{"parsed": time.Now().Format("2006-01-02")},
				sq.Eq{"sent": false},
			},
		).OrderBy("manufacturer", "model")

	rows, err := q.QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	defer rows.Close()
	cars := make([]model.Car, 0)
	for rows.Next() {
		var car model.Car
		if err = rows.Scan(&car.Manufacturer, &car.Model, &car.Year, &car.Mileage, &car.EngineSize, &car.Fuel,
			&car.Drive, &car.AutomaticGearbox, &car.Power, &car.Color, &car.Price, &car.Description, &car.AdID,
			&car.Address, &car.Link, &car.Posted); err != nil {
			return nil, err
		}
		cars = append(cars, car)
	}
	return cars, nil
}

// UpdateSent updates the sent field in the database
func (r *Repository) UpdateSent(ctx context.Context) error {
	q := r.psql.Builder().Update("cars").Set("sent", true).
		Where(sq.Expr("ad_id in (?)",
			sq.Select("ad_id").Distinct().From("cars").Where(sq.Eq{"sent": true})))
	_, err := q.ExecContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.ErrNotFound
		}
	}
	return err
}

// AdSent marks the ad as sent
func (r *Repository) AdSent(ctx context.Context, adId string) error {
	q := r.psql.Builder().Update("cars").Set("sent", true).Where(sq.Eq{"ad_id": adId})
	_, err := q.ExecContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.ErrNotFound
		}
	}
	return err
}

// AdsWithNewPrice marks the ad as sent and updates the price
func (r *Repository) AdsWithNewPrice(ctx context.Context) ([]model.Car, error) {
	q := r.psql.Builder().Select("lc.manufacturer, lc.model, lc.year, lc.mileage, lc.engine, lc.fuel, " +
		"lc.drive, lc.automatic, lc.power, lc.color, lc.price, rc.price as old_price, " +
		"lc.description, lc.ad_id, lc.address, lc.link, lc.posted").
		From("cars as lc").
		Join("cars as rc ON lc.ad_id = rc.ad_id").
		Where(sq.And{
			sq.Expr("rc.price != lc.price"),
			sq.Eq{"lc.manufacturer": favouriteBrands},
			sq.NotEq{"lc.model": excludeModels},
			sq.LtOrEq{"lc.price": maxPrice},
			sq.LtOrEq{"lc.mileage": maxMileage},
			sq.GtOrEq{"lc.year": minYear},
			sq.Eq{"lc.automatic": true},
			sq.GtOrEq{"lc.engine": minEngineSize},
			sq.Eq{"lc.parsed": time.Now().Format("2006-01-02")},
			sq.Eq{"rc.parsed": time.Now().AddDate(0, 0, -1).Format("2006-01-02")},
		})
	rows, err := q.QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	defer rows.Close()
	cars := make([]model.Car, 0)
	for rows.Next() {
		var car model.Car
		if err = rows.Scan(&car.Manufacturer, &car.Model, &car.Year, &car.Mileage, &car.EngineSize, &car.Fuel,
			&car.Drive, &car.AutomaticGearbox, &car.Power, &car.Color, &car.Price, &car.OldPrice, &car.Description, &car.AdID,
			&car.Address, &car.Link, &car.Posted); err != nil {
			return nil, err
		}
		cars = append(cars, car)
	}
	return cars, nil
}

// Close closes the repository
func (r *Repository) Close(ctx context.Context) error {
	return r.psql.Close()
}
