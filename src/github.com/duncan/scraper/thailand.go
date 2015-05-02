package scraper

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/duncan/db"
	"github.com/duncan/weather"
	"strconv"
	"log"
	"fmt"
	"strings"
	"time"
)

//var url string = "http://www.nea.gov.sg/anti-pollution-radiation-protection/air-pollution-control/psi/psi"
var THAILAND_URL string = "http://www.aqmthai.com/index.php?lang=en"

//Return the advisory from NEA
func GetThailandAdvisory(value int) int {
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

func ScrapeThailand() ([]db.City, []ScrapeError) {
	var cities []db.City
	var myFailures []ScrapeError

	doc, err := goquery.NewDocument(THAILAND_URL)
	if err != nil {
		return cities, myFailures
	}

	loc, _ := time.LoadLocation("Asia/Bangkok")
	const longForm = "2006-01-02T15:04:05"

	temp := doc.Find("table#table-body tbody")
	temp.Find("tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return //The first tr is the header of the table
		}
		fmt.Printf("Scraping %-30s #%-4d\r", "Thailand", i - 1)
		tds := s.Find("td")
		city_id := fmt.Sprintf("TH%d", i - 1)
		psi_value := tds.Eq(8).Text()
		//Remove random characters
		psi_value = strings.Replace(psi_value, " ", "", -1)

		city_name := tds.Eq(1).Find("strong a").Text()

		if (len(psi_value) == 0) || (len(city_name) == 0) {
			log.Printf("[THAILAND] Scrape failure: '%s' '%s'\n", psi_value, city_name)
			myFailures = append(myFailures, ScrapeError{city_name, psi_value, "Thailand"})
			return
		}

		date_raw := tds.Eq(2).Text()
		date_raw = strings.Replace(date_raw, " ", "", -1)
		time_raw := tds.Eq(3).Text()
		scrape_time, _ := time.ParseInLocation(longForm, fmt.Sprintf("%sT%s", date_raw, time_raw), loc)

		psi, e1 := strconv.Atoi(psi_value)
		if e1 != nil {
			log.Printf("[THAILAND] Scrape failure: '%s' '%s'\n", psi_value, city_name)
			myFailures = append(myFailures, ScrapeError{city_name, psi_value, "Thailand"})
		} else {
			failsafe := city_name[strings.LastIndex(city_name, ", ") + 2:]
			if len(failsafe) == 0 {
				failsafe = "Thailand"
			}
			th_temp := (int)(weather.GetWeather(city_id, city_name, failsafe).Temp)
			cities = append(cities, db.City{Id: city_id, Name: city_name, Data: psi, Temp: th_temp, AdvisoryCode: GetThailandAdvisory(psi), ScrapeTime: (scrape_time.UnixNano() / 1000000)})
		}
	})

	fmt.Printf("Scraping %-30s Complete\n", "Thailand")
	return cities, myFailures
}