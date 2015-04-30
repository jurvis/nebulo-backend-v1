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

var Updated map[string]bool

var SCRAPE_INTERVAL_MINUTE int = 30

//Return the Unix Epoch in millis
func GetUnixTime() int64 {
	return (time.Now().UnixNano() / 1000000)
}

func DoAlert(failures []string) {
	body := "Hi Masters, I detected some failures:\n\n"

	/*sql_data := "DELETE FROM data WHERE id IN ("
	sql_locations := "DELETE FROM locations WHERE id IN ("
	redis_mass := ""*/

	for _, failure := range failures {
		body += fmt.Sprintf("%s\n", failure)
		/*sql_data += fmt.Sprintf("%d, ", city.Id)
		sql_locations += fmt.Sprintf("%d, ", city.Id)
		redis_mass += fmt.Sprintf("DEL %d\n", city.Id)*/
		fmt.Println(failure)
	}

	/*sql_data = sql_data[:len(sql_data) - 2]
	sql_locations = sql_locations[:len(sql_locations) - 2]

	sql_data += ");"
	sql_locations += ");"*/

	//body += fmt.Sprintf("This is %d out of %d of all the cities I am supposed to scrape.\n", len(failedScrapes), totalCount)
	body += "\nCurrent scrape policy: Cities that scrape successfully are saved, while cities that scrape unsuccessfully retain their last good value. If there has never been a good value, it returns -1.\n"
	body += "\nCurrent push policy: Scraped data must be different from their existing equivalent in the DB and must also be >100 to get push notifications.\n"
	/*body += "\nIf you decide to remove these entries, here are the relevant queries to run:"
	body += "\n\n" + sql_data
	body += "\n\n" + sql_locations
	body += "\n\n" + redis_mass
	body += "\n\nP.S. Run the redis_mass translator to turn it into a suitable pipe-able text file."*/
	body += "\nPlease do check what went wrong. Thanks."

	go email.Alert("Nebulo Backend Scraping Failure!", body)
	fmt.Println("An email was sent to notify of the failures.")
	log.Println("An email was sent to notify of the failures.")
}

func ScrapeInterval() {
	for {
		//Clear counters
		start := time.Now()
		Updated = make(map[string]bool)

		fmt.Println("=====SCRAPER BEGIN=====")
		log.Println("=====SCRAPER BEGIN=====")

		var allCities []db.City
		var failures []string

		weather.ClearCache()

		var TOTAL_COUNTRIES = 4

		//Scrape
		jobChannel := make(chan bool, TOTAL_COUNTRIES) //Total no
		go func() {
			sg := ScrapeSingapore()
			if len(sg) == 0 {
				failures = append(failures, "No Singapore cities were scraped")
			}
			allCities = append(allCities, sg...)
			jobChannel <- true
		}()
		go func() {
			my := ScrapeMalaysia()
			if len(my) == 0 {
				failures = append(failures, "No Malaysia cities were scraped")
			}
			allCities = append(allCities, my...)
			jobChannel <- true
		}()
		go func() {
			hk := ScrapeHongKong()
			if len(hk) == 0 {
				failures = append(failures, "No Hong Kong cities were scraped")
			}
			allCities = append(allCities, hk...)
			jobChannel <- true
		}()
		go func() {
			th := ScrapeThailand()
			if len(th) == 0 {
				failures = append(failures, "No Thailand cities were scraped")
			}
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

		fmt.Println("Transferring Redis temp data into DB.")
		db.SaveData(allCities, Updated)

		fmt.Printf("Scraping complete. Runs again in %d minutes\n", SCRAPE_INTERVAL_MINUTE)
		log.Printf("Scraping complete. Runs again in %d minutes\n", SCRAPE_INTERVAL_MINUTE)
		fmt.Println("=====SCRAPER END=====")

		//Delay for next run
		time.Sleep(time.Duration(SCRAPE_INTERVAL_MINUTE) * time.Minute)
	}
}