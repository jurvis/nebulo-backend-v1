package scraper

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/duncan/db"
	"log"
	"fmt"
	"errors"
)

var AQICN_URL = "http://aqicn.org/city/singapore/central/"

func ScrapeLegacy() (db.LegacyWeather, error) {
	doc, err := goquery.NewDocument("http://aqicn.org/city/singapore/central/")

	if err != nil {
		fmt.Println("Failed to connect to AQICN")
		log.Println(err)
		return db.LegacyWeather{}, errors.New("goquery document creation failed")
	}

	var legacyWeather db.LegacyWeather
	var er error

	doc.Find("div#citydivouter div#citydivmain div.aqiwidget tbody").EachWithBreak(func(i int, s *goquery.Selection) bool {
		psi := s.Find("tr#tr_psi td#cur_psi.tdcur").Text()
		pm25 := s.Find("#tr_pm25 #cur_pm25.tdcur").Text()
		temp := s.Find("#tr_t #cur_t.tdcur").Text()
		fmt.Printf("Legacy Scraper: PSI: %s, PM25: %s Temp: %s\n", psi, pm25, temp)
		log.Printf("Legacy Scraper: PSI: %s, PM25: %s Temp: %s\n", psi, pm25, temp)
		legacyWeather = db.LegacyWeather{PSI: psi, PM25: pm25, Temp: temp}
		if (len(psi) == 0) || (len(pm25) == 0) || (len(temp) == 0) {
			er = errors.New("scrape value empty")
		}
		return false
	})

	fmt.Println("AQICN Scrape complete.")

	if er == nil {
		return legacyWeather, nil
	} else {
		return legacyWeather, er
	}
}