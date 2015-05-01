package scraper

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/duncan/db"
	"github.com/duncan/weather"
	"strconv"
	"log"
	"fmt"
	//"strings"
)

var HONGKONG_URL string = "http://www.aqhi.gov.hk/en/aqhi/pollutant-and-aqhi-distribution.html"

//Return the advisory from EPD
func GetHongKongAdvisory(value int) int {
	if value >= 301 {
		return 4
	} else if value >= 201 {
		return 3
	} else if value >= 101 {
		return 2
	} else if value >= 51 {
		return 1
	} else {
		return 0
	}
}

func ScrapeHongKong() ([]db.City, []ScrapeError){
	var cities []db.City
	var myFailures []ScrapeError

	doc, err := goquery.NewDocument(HONGKONG_URL)
	if err != nil {
		return cities, myFailures
	}

	numGeneralStations := 0

	generalStations := doc.Find("table#tblDistribution_g tbody")

	generalStations.Find("tr").Each(func(i int, s *goquery.Selection) {
		if (i >= 0) && (i <= 2) {
			return //First few trs are headers
		}
		fmt.Printf("Scraping Hong Kong #%d\r", i - 3)
		city_id := fmt.Sprintf("HK%d", i - 3)
		tds := s.Find("td")
		psi_value := tds.Eq(6).Text()
		city_name := tds.Eq(0).Text()

		if (len(psi_value) == 0) || (len(city_name) == 0) {
			log.Printf("[HONG KONG] Scrape failure: '%s' '%s'\n", psi_value, city_name)
			myFailures = append(myFailures, ScrapeError{city_name, psi_value, "Hong Kong"})
			return
		}

		psi, e1 := strconv.ParseFloat(psi_value, 64)
		if e1 != nil {
			log.Printf("[HONG KONG] Scrape failure: '%s' '%s'\n", psi_value, city_name)
			myFailures = append(myFailures, ScrapeError{city_name, psi_value, "Hong Kong"})
		} else {
			hk_temp := (int)(weather.GetWeather(city_id, city_name, "Hong Kong").Temp)
			cities = append(cities, db.City{Id: city_id, Name: city_name, Data: int(psi), Temp: hk_temp, AdvisoryCode: GetHongKongAdvisory(int(psi)), ScrapeTime: GetUnixTime()})
		}
		numGeneralStations++
	})

	roadsideStations := doc.Find("table#tblDistribution_r tbody")

	roadsideStations.Find("tr").Each(func(i int, s *goquery.Selection) {
		if (i >= 0) && (i <= 2) {
			return //First few trs are headers
		}
		fmt.Printf("Scraping Hong Kong #%d\r", i - 3 + numGeneralStations)
		city_id := fmt.Sprintf("HK%d", i - 3 + numGeneralStations)
		tds := s.Find("td")
		psi_value := tds.Eq(6).Text()
		city_name := tds.Eq(0).Text()

		if (len(psi_value) == 0) || (len(city_name) == 0) {
			log.Printf("[HONG KONG] Scrape failure: '%s' '%s'\n", psi_value, city_name)
			myFailures = append(myFailures, ScrapeError{city_name, psi_value, "Hong Kong"})
			return
		}

		psi, e1 := strconv.ParseFloat(psi_value, 64)
		if e1 != nil {
			log.Printf("[HONG KONG] Scrape failure: '%s' '%s'\n", psi_value, city_name)
			myFailures = append(myFailures, ScrapeError{city_name, psi_value, "Hong Kong"})
		} else {
			hk_temp := (int)(weather.GetWeather(city_id, city_name, "Hong Kong").Temp)
			cities = append(cities, db.City{Id: city_id, Name: city_name, Data: int(psi), Temp: hk_temp, AdvisoryCode: GetHongKongAdvisory(int(psi)), ScrapeTime: GetUnixTime()})
		}
	})

	fmt.Println("Scraping Hong Kong Complete")
	return cities, myFailures
}