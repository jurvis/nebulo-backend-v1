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
var MALAYSIA_URL string = "http://apims.doe.gov.my/apims/hourly2.php"

//Return the advisory from NEA
func GetMalaysiaAdvisory(value int) int {
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

func ScrapeMalaysia() ([]db.City, []ScrapeError) {
	var cities []db.City
	var myFailures []ScrapeError

	doc, err := goquery.NewDocument(MALAYSIA_URL)
	if err != nil {
		return cities, myFailures
	}

	list := doc.Find("center div#wrapper div#right table.table1 tbody")

	list.Find("tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return //The first tr is the header of the table
		}
		fmt.Printf("Scraping Malaysia #%d\r", i - 1)
		tds := s.Find("td")
		city_id := fmt.Sprintf("MY%d", i - 1)
		psi_value := tds.Eq(7).Find("font b").Text()
		//Remove random characters
		psi_value = strings.Replace(psi_value, " ", "", -1)
		psi_value = strings.Replace(psi_value, "*", "", -1)
		psi_value = strings.Replace(psi_value, "&", "", -1)
		psi_value = strings.Replace(psi_value, "a", "", -1)
		psi_value = strings.Replace(psi_value, "b", "", -1)
		psi_value = strings.Replace(psi_value, "c", "", -1)
		psi_value = strings.Replace(psi_value, "d", "", -1)

		state := tds.Eq(0).Text()
		city_name := fmt.Sprintf("%s, %s", tds.Eq(1).Text(), state)

		if (len(psi_value) == 0) || (len(city_name) == 0) {
			log.Printf("[MALAYSIA] Scrape failure: '%s' '%s'\n", psi_value, city_name)
			myFailures = append(myFailures, ScrapeError{city_name, psi_value, "Malaysia"})
			return
		}

		psi, e1 := strconv.Atoi(psi_value)
		if e1 != nil {
			log.Printf("[MALAYSIA] Scrape failure: '%s' '%s'\n", psi_value, city_name)
			myFailures = append(myFailures, ScrapeError{city_name, psi_value, "Malaysia"})
		} else {
			my_temp := (int)(weather.GetWeather(city_id, city_name, state).Temp)
			cities = append(cities, db.City{Id: city_id, Name: city_name, Data: psi, Temp: my_temp, AdvisoryCode: GetMalaysiaAdvisory(psi), ScrapeTime: GetUnixTime()})
		}
	})

	fmt.Println("Scraping Malaysia Complete")
	return cities, myFailures
}