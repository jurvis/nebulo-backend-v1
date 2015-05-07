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

func ScrapeThailand(firstIndex int) ([]db.City, []ScrapeError) {
	var cities []db.City
	var myFailures []ScrapeError

	doc, err := goquery.NewDocument(THAILAND_URL)
	if err != nil {
		fmt.Printf("Connect to %-30s: %s\n", "Thailand", "Failed")
		connectFailures = append(connectFailures, ConnectError{"Thailand"})
		return cities, myFailures
	}

	fmt.Printf("Connect to %-30s: %s\n", "Thailand", "Success")

	loc, _ := time.LoadLocation("Asia/Bangkok")
	const longForm = "2006-01-02T15:04:05"

	temp := doc.Find("table#table-body tbody")
	temp.Find("tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i == 0 {
			return true //The first tr is the header of the table
		}
		//Check for second segment which we don't care about
		if s.Find("td strong").Text() == "Reporting of Air Quality Monitoring Stations" {
			return false
		}

		fmt.Printf("Scraping %-30s #%-4d\r", "Thailand", i - 1)
		tds := s.Find("td")
		city_id := firstIndex + (i - 1)
		psi_value := tds.Eq(8).Text()
		//Remove random characters
		psi_value = strings.Replace(psi_value, " ", "", -1)

		city_name := tds.Eq(1).Find("strong a").Text()

		if len(city_name) == 0 {
			return true
		}

		date_raw := tds.Eq(2).Text()
		date_raw = strings.Replace(date_raw, " ", "", -1)
		time_raw := tds.Eq(3).Text()
		scrape_time, _ := time.ParseInLocation(longForm, fmt.Sprintf("%sT%s", date_raw, time_raw), loc)

		psi, e1 := strconv.Atoi(psi_value)

		if (e1 != nil || len(psi_value) == 0) {
			log.Printf("[THAILAND] Scrape failure: '%s' '%s'\n", psi_value, city_name)
			myFailures = append(myFailures, ScrapeError{city_name, psi_value, "Thailand"})
			psi = -1
		}

		failsafe := "Thailand"
		lastCommaIndex := strings.LastIndex(city_name, ", ")
		if lastCommaIndex != -1 {
			failsafe = city_name[strings.LastIndex(city_name, ", ") + 2:]
			//Remove state
			city_name = city_name[:lastCommaIndex]
		}
		th_temp := (int)(weather.GetWeather(city_id, city_name, failsafe).Temp)
		cities = append(cities, db.City{Id: city_id, Name: city_name, Data: psi, Temp: th_temp, AdvisoryCode: GetThailandAdvisory(psi), ScrapeTime: (scrape_time.UnixNano() / 1000000)})
		return true
	})

	fmt.Printf("Scraping %-30s Complete\n", "Thailand")
	return cities, myFailures
}