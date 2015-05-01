package scraper

import (
	"fmt"
	"log"
	"time"
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

var failures []ScrapeError

var SCRAPE_INTERVAL_MINUTE int = 30

//Return the Unix Epoch in millis
func GetUnixTime() int64 {
	return (time.Now().UnixNano() / 1000000)
}

func DoAlert() {
	body := "Hi Masters, I detected some failures:\n\n"

	for _, failure := range failures {
		body += fmt.Sprintf("%s city '%s' : %s\n", failure.Country, failure.City_name, failure.Data)
	}

	body += "\nCurrent scrape policy: Cities that scrape successfully are saved, while cities that scrape unsuccessfully retain their last good value. If there has never been a good value, it returns -1.\n"
	body += "\nCurrent push policy: Scraped data must be different from their existing equivalent in the DB and must also be >100 to get push notifications.\n"
	body += "\nPlease do check what went wrong. Thanks."

	go email.Alert("Nebulo Backend Scraping Failure!", body)
	fmt.Println("An email was sent to notify of the failures.")
	log.Println("An email was sent to notify of the failures.")
}

func ScrapeInterval() {
	for {
		//Clear counters
		start := time.Now()

		fmt.Println("=====SCRAPER BEGIN=====")
		log.Println("=====SCRAPER BEGIN=====")

		var allCities []db.City

		//Clear failures
		failures = nil

		weather.ClearCache()

		var TOTAL_COUNTRIES = 4

		//Scrape
		jobChannel := make(chan bool, TOTAL_COUNTRIES) //Total no
		go func() {
			sg, fail := ScrapeSingapore()
			failures = append(failures, fail...)
			allCities = append(allCities, sg...)
			jobChannel <- true
		}()
		go func() {
			my, fail := ScrapeMalaysia()
			failures = append(failures, fail...)
			allCities = append(allCities, my...)
			jobChannel <- true
		}()
		go func() {
			hk, fail := ScrapeHongKong()
			failures = append(failures, fail...)
			allCities = append(allCities, hk...)
			jobChannel <- true
		}()
		go func() {
			th, fail := ScrapeThailand()
			failures = append(failures, fail...)
			allCities = append(allCities, th...)
			jobChannel <- true
		}()

		for i := 0; i < TOTAL_COUNTRIES; i++ { //Total no
			<-jobChannel
		}

		for _, city := range allCities {
			push.Push(city)
		}

		end := time.Now()
		timeElapsedMillis := int64(end.Sub(start) / time.Nanosecond) / 1000000

		fmt.Printf("Scraping complete. %d cities were successfully scraped.\n", len(allCities))
		log.Printf("Scraping complete. %d cities were successfully scraped.\n", len(allCities))
		fmt.Printf("Duration elapsed: %dms.\n", timeElapsedMillis)

		if len(failures) > 0 {
			DoAlert()
		}
		fmt.Println("Saving data")
		db.SaveData(allCities)
		fmt.Println("Saved data")

		fmt.Printf("Scraping complete. Runs again in %d minutes\n", SCRAPE_INTERVAL_MINUTE)
		log.Printf("Scraping complete. Runs again in %d minutes\n", SCRAPE_INTERVAL_MINUTE)
		fmt.Println("=====SCRAPER END=====")

		//Delay for next run
		time.Sleep(time.Duration(SCRAPE_INTERVAL_MINUTE) * time.Minute)
	}
}