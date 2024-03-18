package parser

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bopoh24/bazacars/internal/model"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ParseCarPage parses the page and returns the car data
func ParseCarPage(url string) (model.Car, error) {
	res, err := http.Get(url)
	if err != nil {
		return model.Car{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return model.Car{}, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	carData, err := extractCarData(res.Body)
	if err != nil {
		return model.Car{}, err
	}
	carData.Link = url
	return carData, nil
}

func extractCarData(body io.ReadCloser) (model.Car, error) {
	var carData model.Car
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(body)

	if err != nil {
		return carData, err
	}
	modelLi := doc.Find(".breadcrumbs li").Last()
	carData.Model = modelLi.Find("span").Text()
	manufacturerLi := modelLi.Prev()
	carData.Manufacturer = manufacturerLi.Find("span").Text()

	postedTxt := doc.Find(".date-meta").Text()
	postedTxt = strings.TrimLeft(postedTxt, "Posted: ")

	carData.Posted, err = time.Parse("02.01.2006 15:04", postedTxt)
	if err != nil {
		return carData, err
	}
	carData.AdID = doc.Find(".number-announcement span").Text()
	carData.Address = doc.Find(".announcement__location span").Text()

	// parse car characteristics
	doc.Find(".chars-column .key-chars").Each(func(i int, s *goquery.Selection) {
		switch s.Text() {
		// year
		case "Year:":
			yearText := s.Next().Text()
			carData.Year, err = strconv.Atoi(yearText)
			if err != nil {
				slog.Error("Error parsing year", "year", yearText, "err", err)
			}
		// wheel drive
		case "Drive:":
			carData.Drive = parseDriveType(s.Next().Text())

		// power
		case "Power:":
			powerTxt := s.Next().Text()
			powerTxt = strings.TrimRight(powerTxt, " hp")
			carData.Power, err = strconv.Atoi(powerTxt)
			if err != nil {
				slog.Error("Error parsing power", "power", powerTxt, "err", err)
			}
		// color
		case "Colour:":
			carData.Color = strings.ToLower(s.Next().Text())

		// gearbox
		case "Gearbox:":
			carData.AutomaticGearbox = s.Next().Text() == "Automatic"

		// mileage
		case "Mileage (in km):":
			mileageTxt := strings.TrimRight(s.Next().Text(), " km")
			carData.Mileage, err = strconv.Atoi(mileageTxt)

		// fuel type
		case "Fuel type:":
			carData.Fuel = model.FuelType(s.Next().Text())

		case "Engine size:":
			engineSizeTxt := strings.TrimRight(s.Next().Text(), "L")
			engineSizeTxt = strings.Replace(engineSizeTxt, ",", ".", -1)
			carData.EngineSize, err = strconv.ParseFloat(engineSizeTxt, 64)
			if err != nil {
				slog.Error("Error parsing engine size", "engine_size", engineSizeTxt, "err", err)
			}
		}

	})
	descriptionText := doc.Find(".js-description").Text()
	descriptionText = strings.Replace(descriptionText, "<p>", "", -1)
	descriptionText = strings.Replace(descriptionText, "</p>", "\n", -1)
	descriptionText = strings.TrimSpace(descriptionText)
	carData.Description = descriptionText

	// parse price
	doc.Find(".announcement-price__cost meta").Each(func(i int, s *goquery.Selection) {
		val, ok := s.Attr("itemprop")
		if ok && val == "price" {
			priceText, _ := s.Attr("content")
			split := strings.Split(priceText, ".")
			priceText = split[0]
			carData.Price, err = strconv.Atoi(priceText)
			if err != nil {
				slog.Error("Error parsing price", "price", priceText, "err", err)
			}
		}
	})
	return carData, nil
}

func parseDriveType(drive string) model.DriveType {
	if strings.Contains(drive, string(model.DriveTypeAll)) {
		return model.DriveTypeAll
	}
	if strings.Contains(drive, string(model.DriveTypeRear)) {
		return model.DriveTypeRear
	}
	return model.DriveTypeFront
}
