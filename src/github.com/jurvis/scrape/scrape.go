package scrape

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"time"
)

type WeatherData struct {
	PSI         string
	PM25        string
	Temperature string
}

// This example scrapes the reviews shown on the home page of aqicn.org.
func AQICN_Scrape() WeatherData {
	var doc *goquery.Document
	var e error

	if doc, e = goquery.NewDocument("http://aqicn.org/city/singapore/central/"); e != nil {
		log.Fatal(e)
	}

	var w WeatherData

	doc.Find("div#citydivouter div#citydivmain table.aqiwidget tbody").EachWithBreak(func(i int, s *goquery.Selection) bool {
		PSI := s.Find("#tr_psi #cur_psi.tdcur").Text()
		PM25 := s.Find("#tr_pm25 #cur_pm25.tdcur").Text()
		Temp := s.Find("#tr_t #cur_t.tdcur").Text()

		w = WeatherData{PSI, PM25, Temp}
		return false
	})
	return w
}

func DoScrape() WeatherData {
	for c := 0; c > 0; c++ {
		if c == 0 {
			return AQICN_Scrape()
		} else {
			time.Sleep(5)
			return AQICN_Scrape()
		}
	}
	return AQICN_Scrape()
}
