package scraper

import (
	"fmt"
	"log"
	"time"
	"strings"
	"github.com/duncan/db"
	"github.com/duncan/email"
	"github.com/duncan/push"
	"github.com/duncan/weather"
)

type ScrapeError struct {
	City_name string
	Data string
	Country string
}

type ConnectError struct {
	Country string
}

var failures []ScrapeError
var connectFailures []ConnectError

var SCRAPE_INTERVAL_MINUTE int = 30

var BLACKLIST []string = []string{"Residence for Dept. of Primary Industries and Mines, Samut Prakan"}

//Return the Unix Epoch in millis
func GetUnixTime() int64 {
	return (time.Now().UnixNano() / 1000000)
}

func DoAlert() {
	body := "FAILURES\n"
	//body := "Hi Masters, I detected some failures:\n\n"

	for _, connectFailure := range connectFailures {
		body += fmt.Sprintf("%s\n==============\nConnection Failed. Site may be down. Nothing scraped.\n", connectFailure.Country)
	}

	country := ""
	for _, failure := range failures {
		if failure.Country != country {
			country = failure.Country
			body += fmt.Sprintf("\n\n%s\n==============\n", country)
		}
		body += fmt.Sprintf("%s: '%s'\n", failure.City_name, failure.Data)
	}

	body += "\n\nCURRENT POLICIES\n"
	body += "\nScrape\nCities that scrape successfully are saved, while cities that scrape unsuccessfully retain their last good value. If there has never been a good value, it returns -1.\n"
	body += "\nAPI\nCities with Data: -1 or Temp: 99999 (the pre-defined invalid temp) are NOT RETURNED.\n"
	body += "\nPush\nScraped data must be different from their existing equivalent in the DB and must also be >100 to get push notifications.\n"
	//body += "\nPlease do check what went wrong. Thanks."

	go email.Alert("Nebulo Backend Scraping Failure!", body)
	fmt.Println("An email was sent to notify of the failures.")
	log.Println("An email was sent to notify of the failures.")
}

func ScrapeInterval() {
	for {
		//Clear counters
		start := time.Now()

		/*fmt.Println("=====LEGACY SCRAPE=====")
		log.Println("=====LEGACY SCRAPE=====")

		legacyWeather, err := ScrapeLegacy()
		if err == nil {
			db.SaveLegacyData(legacyWeather)
			fmt.Println("Saved legacy data\n")
			log.Println("Saved legacy data\n")
		} else {
			log.Println(err)
		}*/

		//Legacy scrape is currently an AVG function on the five Singapore cities.

		fmt.Println("=====SCRAPER START=====")
		log.Println("=====SCRAPER START=====")

		var allCities []db.City

		//Clear failures
		failures = nil
		connectFailures = nil

		weather.ClearCache()

		var count int = 0

		//Scrape
		sg, sg_fail := ScrapeSingapore(count)
		failures = append(failures, sg_fail...)
		allCities = append(allCities, sg...)

		count += len(sg)

		hk, hk_fail := ScrapeHongKong(count)
		failures = append(failures, hk_fail...)
		allCities = append(allCities, hk...)

		count += len(hk)

		my, my_fail := ScrapeMalaysia(count)
		failures = append(failures, my_fail...)
		allCities = append(allCities, my...)

		count += len(my)

		th, th_fail := ScrapeThailand(count)
		failures = append(failures, th_fail...)
		allCities = append(allCities, th...)

		count += len(th)

		for _, city := range allCities {
			push.Push(city)
		}

		end := time.Now()
		timeElapsedMillis := int64(end.Sub(start) / time.Nanosecond) / 1000000

		fmt.Printf("Scraping complete. %d cities were successfully scraped.\n", len(allCities))
		log.Printf("Scraping complete. %d cities were successfully scraped.\n", len(allCities))
		fmt.Printf("Duration elapsed: %dms.\n", timeElapsedMillis)

		if (len(failures) > 0 || len(connectFailures) > 0) {
			DoAlert()
		}
		fmt.Println("Saving data")
		db.SaveData(GarbageCollection(allCities))
		fmt.Println("Saved data")

		fmt.Printf("Scraping complete. Runs again in %d minutes\n", SCRAPE_INTERVAL_MINUTE)
		log.Printf("Scraping complete. Runs again in %d minutes\n", SCRAPE_INTERVAL_MINUTE)
		fmt.Println("=====SCRAPER END=====")

		//Delay for next run
		time.Sleep(time.Duration(SCRAPE_INTERVAL_MINUTE) * time.Minute)
	}
}

func contains(s []string, e string) bool {
    for _, a := range s {
    	if strings.Contains(a, e) {
    		return true
    	}
    }
    return false
}

func GarbageCollection(allCities []db.City) []db.City {
	var ret []db.City
	for _, city := range allCities {
		if (!contains(BLACKLIST, city.Name)) {
			//Delete it
			ret = append(ret, city)
		}
	}
	return ret
}