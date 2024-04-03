package repository

import (
	"context"
	"github.com/bopoh24/bazacars/internal/model"
)

type Repository interface {
	SaveCars(ctx context.Context, car []model.Car) error
	NewAds(ctx context.Context) ([]model.Car, error)
	AdSent(ctx context.Context, adId string) error
	AdsWithNewPrice(ctx context.Context) ([]model.Car, error)
	Users(ctx context.Context) ([]model.User, error)
	User(ctx context.Context, chatID int64) (model.User, error)
	Admins(ctx context.Context) ([]model.User, error)
	UserAdd(ctx context.Context, user model.User) error
	UserSave(ctx context.Context, user model.User) error
	Close(ctx context.Context) error
}
