package model

import (
	"fmt"
	"time"
)

type DriveType string

type FuelType string

const (
	DriveTypeFront DriveType = "FWD"
	DriveTypeRear  DriveType = "RWD"
	DriveTypeAll   DriveType = "AWD"

	FuelTypeDiesel             FuelType = "Diesel"
	FuelTypePetrol             FuelType = "Petrol"
	FuelTypePluginHybridDiesel FuelType = "Plug-In Hybrid Diesel"
	FuelTypePluginHybridPetrol FuelType = "Plug-In Hybrid Petrol"
	FuelTypeHybridDiesel       FuelType = "Hybrid Diesel"
	FuelTypeHybridPetrol       FuelType = "Hybrid Petrol"
	FuelTypeElectric           FuelType = "Electric"
	FuelTypeLPG                FuelType = "LPG"
)

type Car struct {
	Manufacturer     string    `json:"manufacturer"`
	Model            string    `json:"model"`
	Year             int       `json:"year"`
	Mileage          int       `json:"mileage"`
	EngineSize       float64   `json:"engine"`
	Fuel             FuelType  `json:"fuel"`
	Drive            DriveType `json:"drive"`
	AutomaticGearbox bool      `json:"automatic"`
	Power            int       `json:"power"`
	Color            string    `json:"color"`
	Price            int       `json:"price"`
	OldPrice         int       `json:"old_price,omitempty"`
	Description      string    `json:"description"`
	AdID             string    `json:"ad_id"`
	Link             string    `json:"link"`
	Posted           time.Time `json:"posted"`
	Address          string    `json:"address"`
	Parsed           time.Time `json:"parsed"`
	Sent             bool      `json:"sent"`
}

// User model with chatID
type User struct {
	ChatID    int64
	FirstName string
	LastName  string
	Username  string
	Admin     bool
	Approved  bool
	UpdatedAt time.Time
	CreatedAt time.Time
}

func (u User) String() string {
	result := ""
	if u.Username != "" {
		result += "[@" + u.Username + "] "
	}
	result += fmt.Sprintf("%s %s", u.FirstName, u.LastName)
	return result
}
