package scrape

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/steveyen/gkvlite"
	"log"
	"os"
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

func storeData() {
	fmt.Println("Scraping...")
	w := AQICN_Scrape()
	file, err := os.Create("/tmp/test.gkvlite")
	if err != nil {
		fmt.Println("Unable to create .gkvlite file")
	}
	s, err := gkvlite.NewStore(file)
	c := s.SetCollection("weatherData", nil)

	c.Set([]byte("PSI"), []byte(w.PSI))
	c.Set([]byte("PM25"), []byte(w.PM25))
	c.Set([]byte("Temp"), []byte(w.Temperature))

	s.Flush()
}

func GetData(d string) string {
	if _, err := os.Stat("/tmp/test.gkvlite"); os.IsNotExist(err) {
		storeData()
	}
	f2, err := os.Open("/tmp/test.gkvlite")
	s2, err := gkvlite.NewStore(f2)
	c2 := s2.GetCollection("weatherData")

	PSI, err := c2.Get([]byte("PSI"))
	PM25, err := c2.Get([]byte("PM25"))
	Temp, err := c2.Get([]byte("Temp"))
	fmt.Println(err)
	switch d {
	case "PSI":
		return string(PSI)
	case "PM25":
		return string(PM25)
	case "Temp":
		return string(Temp)
	default:
		return "not valid"
	}
}
