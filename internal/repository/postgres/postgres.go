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
		if err = rows.Scan(&user.ChatID, &user.FirstName, &user.LastName, &user.Username, &user.Admin, &user.Approved, &user.UpdatedAt, &user.CreatedAt); err != nil {
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
	if err := row.Scan(&user.ChatID, &user.FirstName, &user.LastName, &user.Username, &user.Admin, &user.Approved, &user.UpdatedAt, &user.CreatedAt); err != nil {
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
		err = rows.Scan(&user.ChatID, &user.FirstName, &user.LastName, &user.Username, &user.Admin,
			&user.Approved, &user.UpdatedAt, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *Repository) UserAdd(ctx context.Context, user model.User) error {
	q := r.psql.Builder().Insert("users").Columns("chat_id", "first_name", "last_name", "username", "admin", "approved", "updated_at", "created_at")
	q = q.Values(user.ChatID, user.FirstName, user.LastName, user.Username, user.Admin, user.Approved, user.UpdatedAt, user.CreatedAt)
	_, err := q.ExecContext(ctx)
	return err
}

func (r *Repository) UserSave(ctx context.Context, user model.User) error {
	q := r.psql.Builder().Update("users").
		Set("first_name", user.FirstName).
		Set("last_name", user.LastName).
		Set("username", user.Username).
		Set("admin", user.Admin).
		Set("approved", user.Approved).
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

// Close closes the repository
func (r *Repository) Close(ctx context.Context) error {
	return r.psql.Close()
}
