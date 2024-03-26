package parser

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bopoh24/bazacars/internal/model"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/119.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3.1 Safari/605.1.15",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14.4; rv:124.0) Gecko/20100101 Firefox/124.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
}

func forwardIP() string {
	return fmt.Sprintf("31.216.%d.%d", 64+rand.Intn(255-64), 1+rand.Intn(253))
}

// ParseCarBrands parses the page and returns the list of car manufacturers
func ParseCarBrands(url string) (map[string]string, error) {
	body, err := urlBody(url)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return extractCarManufacturers(body)
}

// ParseAdList parses the page and returns the list of ads
func ParseAdList(url string) ([]string, error) {
	body, err := urlBody(url)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return extractAdList(body)
}

// ParseCarPage parses the page and returns the car data
func ParseCarPage(url string) (model.Car, error) {
	body, err := urlBody(url)
	if err != nil {
		return model.Car{}, err
	}
	defer body.Close()
	carData, err := extractCarData(body)
	if err != nil {
		return model.Car{}, err
	}
	carData.Link = url
	return carData, nil
}

// TotalPages returns the total number of pages
func TotalPages(url string) (int, error) {
	body, err := urlBody(url)
	if err != nil {
		return 0, err
	}
	defer body.Close()
	return extractTotalPages(body)
}

func extractTotalPages(body io.ReadCloser) (int, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return 0, err
	}
	pages := doc.Find(".number-list a").Last().Text()
	if pages == "" {
		return 1, nil
	}
	return strconv.Atoi(pages)
}

func urlBody(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])
	req.Header.Set("X-Forwarded-For", forwardIP())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrStatusNotFound
		}
		if resp.StatusCode == http.StatusForbidden {
			return nil, ErrStatusForbidden
		}
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}
	return resp.Body, nil
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

	postedTxt := doc.Find(".announcement__details .date-meta").Text()
	postedTxt = strings.TrimLeft(postedTxt, "Posted: ")
	if strings.HasPrefix(postedTxt, "Yesterday ") {
		postedTxt = strings.TrimLeft(postedTxt, "Yesterday ")
		postedTxt = time.Now().AddDate(0, 0, -1).Format("02.01.2006 ") + postedTxt
	}
	if strings.HasPrefix(postedTxt, "Today ") {
		postedTxt = strings.TrimLeft(postedTxt, "Today ")
		postedTxt = time.Now().Format("02.01.2006 ") + postedTxt
	}

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

func extractCarManufacturers(body io.ReadCloser) (map[string]string, error) {
	result := make(map[string]string)
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}
	doc.Find(".rubrics-list.toggle-content a").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("href")
		brand := strings.TrimSuffix(s.Text(), "&nbsp;")
		brand = strings.TrimSpace(brand)
		result[brand] = link
	})
	return result, nil
}

func extractAdList(body io.ReadCloser) ([]string, error) {
	var result []string
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}
	doc.Find(".advert__body").Each(func(i int, s *goquery.Selection) {
		a := s.Find("a").First()
		link, _ := a.Attr("href")
		parsedURL, err := url.Parse(link)
		if err != nil {
			return
		}
		// Remove query parameters
		parsedURL.RawQuery = ""
		result = append(result, parsedURL.String())
	})
	return result, nil
}
