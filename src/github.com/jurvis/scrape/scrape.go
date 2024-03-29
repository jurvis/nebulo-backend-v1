package scrape

import (
	"github.com/PuerkitoBio/goquery"
	"log"
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

	doc.Find("div#citydivouter div#citydivmain div.aqiwidget tbody").EachWithBreak(func(i int, s *goquery.Selection) bool {
		PSI := s.Find("tr#tr_psi td#cur_psi.tdcur").Text()
		PM25 := s.Find("#tr_pm25 #cur_pm25.tdcur").Text()
		Temp := s.Find("#tr_t #cur_t.tdcur").Text()
		log.Printf("Scrapped: PSI: %s, PM25: %s Temp: %s", PSI, PM25, Temp)
		w = WeatherData{PSI, PM25, Temp}
		return false
	})
	return w
}
