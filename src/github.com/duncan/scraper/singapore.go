package scraper

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/duncan/db"
	"github.com/duncan/weather"
	"strconv"
	"log"
	"fmt"
	"strings"
)

//var url string = "http://www.nea.gov.sg/anti-pollution-radiation-protection/air-pollution-control/psi/psi"
var SINGAPORE_URL string = "http://www.haze.gov.sg/onemappsi/"

//Return the advisory from NEA
func GetSingaporeAdvisory(value int) int {
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

func ScrapeSingapore() ([]db.City, []ScrapeError) {
	var cities []db.City
	var myFailures []ScrapeError

	doc, err := goquery.NewDocument(SINGAPORE_URL)
	if err != nil {
		return cities, myFailures
	}


	list := doc.Find("ul.list")

	sg_temp := (int)(weather.GetWeather("SG0", "Singapore", "Singapore").Temp)

	list.Children().Each(func(i int, s *goquery.Selection) {
		fmt.Printf("Scraping Singapore #%d\r", i)
		city_id := fmt.Sprintf("SG%d", i)
		psi_value := s.Find("span.psi-value").Text()
		//Remove random characters
		psi_value = strings.Replace(psi_value, " ", "", -1)
		psi_value = strings.Replace(psi_value, "\n", "", -1)

		direction := s.Find("span.direction").Text() //e.g. North
		city_name := fmt.Sprintf("%s, Singapore", direction)

		if (len(psi_value) == 0) || (len(direction) == 0) {
			log.Printf("[SINGAPORE] Scrape failure: '%s' '%s'\n", psi_value, direction)
			myFailures = append(myFailures, ScrapeError{city_name, psi_value, "Singapore"})
			return
		}

		psi, e1 := strconv.Atoi(psi_value)
		if e1 != nil {
			log.Printf("[SINGAPORE] Scrape failure: '%s' '%s'\n", psi_value, direction)
			myFailures = append(myFailures, ScrapeError{city_name, psi_value, "Singapore"})
		} else {
			cities = append(cities, db.City{Id: city_id, Name: city_name, Data: psi, Temp: sg_temp, AdvisoryCode: GetSingaporeAdvisory(psi), ScrapeTime: GetUnixTime()})
		}
	})

	fmt.Println("Scraping Singapore Complete")

	return cities, myFailures
}