package parser

import (
	"github.com/bopoh24/bazacars/internal/model"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestExtractCarData(t *testing.T) {
	f, err := os.Open("testing/item.sample")
	assert.NoError(t, err)

	carData, err := extractCarData(f)
	assert.NoError(t, err)

	assert.Equal(t, "5080505", carData.AdID)
	assert.Equal(t, "Paphos, Geroskipou", carData.Address)
	assert.Equal(t, "24.02.2024 14:06", carData.Posted.Format("02.01.2006 15:04"))
	assert.Equal(t, 2020, carData.Year)
	assert.Equal(t, model.DriveTypeRear, carData.Drive)
	assert.Equal(t, 150, carData.Power)
	assert.Equal(t, "white", carData.Color)
	assert.True(t, carData.AutomaticGearbox)
	assert.Equal(t, 41051, carData.Mileage)
	assert.Equal(t, model.FuelTypePetrol, carData.Fuel)
	assert.Equal(t, 2.0, carData.EngineSize)
	assert.NotEmpty(t, carData.Description)
	assert.Equal(t, "BMW", carData.Manufacturer)
	assert.Equal(t, "3-Series", carData.Model)
	assert.Equal(t, 25900, carData.Price)
}

func TestExtractCarManufactures(t *testing.T) {
	f, err := os.Open("testing/main.sample")
	assert.NoError(t, err)
	result, err := extractCarManufacturers(f)
	assert.NoError(t, err)
	assert.Contains(t, result, "Audi")
	assert.Contains(t, result, "BMW")
	assert.Contains(t, result, "Mercedes-Benz")
	assert.Contains(t, result, "Porsche")
	assert.Contains(t, result, "Volkswagen")
	assert.Equal(t, result["Audi"], "/car-motorbikes-boats-and-parts/cars-trucks-and-vans/audi/")
}

func TestExtractAdList(t *testing.T) {

	f, err := os.Open("testing/main.sample")
	assert.NoError(t, err)
	result, err := extractAdList(f)
	assert.NoError(t, err)
	assert.Equal(t, 60, len(result))
	assert.Equal(t, "/adv/4949445_suzuki-sx4-1-6l-2019/", result[0])
}

func TestExtractTotalPages(t *testing.T) {
	f, err := os.Open("testing/main.sample")
	assert.NoError(t, err)
	result, err := extractTotalPages(f)
	assert.NoError(t, err)
	assert.Equal(t, 154, result)
}
