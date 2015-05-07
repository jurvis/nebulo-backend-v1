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
var MALAYSIA_URL string = "http://apims.doe.gov.my/apims/hourly%d.php"

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

//Return page url
func GetPageUrl() string {
	return fmt.Sprintf(MALAYSIA_URL, GetIndex())
}

//Return index based on time of day (APIMS has 4 pages of data)
func GetIndex() int {
	hour := time.Now().Hour()
	if hour >= 18 {
		return 4
	} else if hour >= 12 {
		return 3
	} else if hour >= 6 {
		return 2
	} else {
		return 1
	}
}

func CleanData(orig string) string {
	orig = strings.Replace(orig, " ", "", -1)
	orig = strings.Replace(orig, "*", "", -1)
	orig = strings.Replace(orig, "&", "", -1)
	orig = strings.Replace(orig, "a", "", -1)
	orig = strings.Replace(orig, "b", "", -1)
	orig = strings.Replace(orig, "c", "", -1)
	orig = strings.Replace(orig, "d", "", -1)
	return orig
}

func ScrapeMalaysia(firstIndex int) ([]db.City, []ScrapeError) {
	var cities []db.City
	var myFailures []ScrapeError

	doc, err := goquery.NewDocument(GetPageUrl())
	if err != nil {
		fmt.Printf("Connect to %-30s: %s\n", "Malaysia", "Failed")
		connectFailures = append(connectFailures, ConnectError{"Malaysia"})
		return cities, myFailures
	}

	fmt.Printf("Connect to %-30s: %s\n", "Malaysia", "Success")

	list := doc.Find("center div#wrapper div#right table.table1 tbody")

	list.Find("tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return //The first tr is the header/times of the table
		}
		fmt.Printf("Scraping %-30s #%-4d\r", "Malaysia", i - 1)
		tds := s.Find("td")
		state := tds.Eq(0).Text()
		city_id := firstIndex + (i - 1)
		//city_name := fmt.Sprintf("%s, %s", tds.Eq(1).Text(), state)
		city_name := tds.Eq(1).Text()
		psi_value := ""
		now := time.Now()
		data_collect_hour := now.Hour()

		if len(city_name) == 0 {
			return
		}

		for i := 7; i >= 2; i-- {
			psi_value = CleanData(tds.Eq(i).Find("font b").Text())
			if len(psi_value) == 0 {
				continue
			}

			_, parse_err := strconv.Atoi(psi_value)
			if parse_err == nil {
				data_collect_hour = (GetIndex() - 1) * 6 + (i - 2)
				break
			}
		}

		scrape_time := time.Date(now.Year(), now.Month(), now.Day(), data_collect_hour, 0, 0, 0, now.Location())
		scrape_time_millis := scrape_time.UnixNano() / 1000000

		psi, e1 := strconv.Atoi(psi_value)

		if (e1 != nil || len(psi_value) == 0) {
			log.Printf("[MALAYSIA] Scrape failure: '%s' '%s'\n", psi_value, city_name)
			myFailures = append(myFailures, ScrapeError{city_name, psi_value, "Malaysia"})
			psi = -1
		}

		my_temp := (int)(weather.GetWeather(city_id, city_name, state).Temp)
		cities = append(cities, db.City{Id: city_id, Name: city_name, Data: psi, Temp: my_temp, AdvisoryCode: GetMalaysiaAdvisory(psi), ScrapeTime: scrape_time_millis})
	})
	
	fmt.Printf("Scraping %-30s Complete\n", "Malaysia")
	return cities, myFailures
}