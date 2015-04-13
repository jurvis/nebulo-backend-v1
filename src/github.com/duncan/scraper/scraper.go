package scraper

import (
	"fmt"
	"log"
	"time"
	"strconv"
	"github.com/PuerkitoBio/goquery"
	"github.com/duncan/db"
	"github.com/duncan/push"
	"github.com/duncan/email"
)

var failCount, totalCount int = 0, 0
var failedScrapes []string = make([]string, 0)

var SCRAPE_INTERVAL_MINUTE, SCRAPE_EACH_INTERVAL_NANOSECOND int = 30, 1e+8

func ReflectFailure(url string) {
	failCount++
	failedScrapes = append(failedScrapes, url)
}

func ScrapeInterval() {
	for {
		//Clear counters
		start := time.Now()
		failedScrapes = failedScrapes[0:0]
		failCount = 0
		totalIntervalNano := 0

		fmt.Println("=====SCRAPER BEGIN=====")
		var cities []db.CityURL = db.GetAllCityURLs()
		totalCount = len(cities)

		fmt.Printf("Scraping %d cities with an interval of %dms between each scrape.\n", len(cities), SCRAPE_EACH_INTERVAL_NANOSECOND / 1000000)

		for _, city := range cities {
			//fmt.Println("Scraping: ", city.Url)
			Scrape(city.Id, city.Url)
			time.Sleep(time.Duration(SCRAPE_EACH_INTERVAL_NANOSECOND) * time.Nanosecond)
			totalIntervalNano += SCRAPE_EACH_INTERVAL_NANOSECOND
		}

		end := time.Now()
		timeElapsedNano := int64(end.Sub(start) / time.Nanosecond) - int64(totalIntervalNano)

		fmt.Printf("Scraping complete. %d cities were successfully scraped.\n", totalCount - failCount)
		fmt.Printf("Duration elapsed: %dms. (Excluding intervals between scrapes)\n", timeElapsedNano / 1000000)

		fmt.Println("Transferring Redis temp data into DB.")
		db.SaveRedisDataIntoDB()

		if failCount > 0 {
			fmt.Printf("There were %d failures:\n", failCount)
			body := "Hi Masters, I detected a failure in scraping the below URLs:\n\n"
			for _, url := range failedScrapes {
				body += url + "\n"
				fmt.Println(url)
			}
			body += "\nPlease do check what went wrong. Thanks."
			go email.Alert("Nebulo Backend Scraping Failure!", body)
			fmt.Println("An email was sent to notify of the failures.")
		}

		fmt.Println("=====SCRAPER END=====")

		//Delay for next run
		time.Sleep(time.Duration(SCRAPE_INTERVAL_MINUTE) * time.Minute)
	}
}

func Scrape(id int, url string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
		ReflectFailure(url)
	}

	finder := doc.Find("div#citydivouter div#citydivmain div.aqiwidget tbody")

	if finder.Length() == 0 {
		ReflectFailure(url)
		return
	}

	finder.EachWithBreak(func(i int, s *goquery.Selection) bool {
		pm25_string := s.Find("tr#tr_pm25 td#cur_pm25.tdcur").Text()
		aqi_string := s.Find("tr#tr_aqi td#cur_aqi.tdcur").Text()
		temp_string := s.Find("tr#tr_t td#cur_t.tdcur").Text()

		temp, err := strconv.Atoi(temp_string)
		if err != nil {
			return false
		}
		
		if len(pm25_string) != 0 {
			pm25, err := strconv.Atoi(pm25_string)
			if err != nil {
				ReflectFailure(url)
				return false
			}
			db.SaveLocationResult(id, pm25, temp)
			push.Push(id, pm25)
		} else if len(aqi_string) != 0 {
			aqi, err := strconv.Atoi(aqi_string)
			if err != nil {
				ReflectFailure(url)
				return false
			}
			db.SaveLocationResult(id, aqi, temp)
			push.Push(id, aqi)
		} else {
			ReflectFailure(url)
		}
		return false
	})
}