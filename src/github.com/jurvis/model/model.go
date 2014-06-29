package model

import (
	"github.com/jurvis/scrape"
	"github.com/steveyen/gkvlite"
	"log"
	"os"
	"strconv"
	"time"
)

func storeData() {
	log.Println("Scraping...")
	w := scrape.AQICN_Scrape()
	file, err := os.Create("/tmp/test.gkvlite")
	if err != nil {
		log.Println("Unable to create .gkvlite file")
	}
	s, err := gkvlite.NewStore(file)
	c := s.SetCollection("weatherData", nil)

	c.Set([]byte("PSI"), []byte(w.PSI))
	c.Set([]byte("PM25"), []byte(w.PM25))
	c.Set([]byte("Temp"), []byte(w.Temperature))

	s.Flush()
}

func checkWeather(pm25 string) string {
	int_pm25, err := strconv.Atoi(pm25)
	if err != nil {
		log.Println("unable to convert string")
	}

	var status string
	if int_pm25 > 200 {
		status = "Stay Indoors."
	} else if int_pm25 > 100 {
		status = "The Air Is Bad."
	} else if int_pm25 > 50 {
		status = "Moderate."
	} else {
		status = "It's Clear."
	}

	return status
}

func RetrieveData(d string) string {
	// check if file exists or if file is more than an hour old (to refresh data)
	if info, err := os.Stat("/tmp/test.gkvlite"); os.IsNotExist(err) || time.Since(info.ModTime()).Hours() > 1 {
		storeData()
	}

	f2, err := os.Open("/tmp/test.gkvlite")
	s2, err := gkvlite.NewStore(f2)
	c2 := s2.GetCollection("weatherData")

	PSI, err := c2.Get([]byte("PSI"))
	PM25, err := c2.Get([]byte("PM25"))
	Temp, err := c2.Get([]byte("Temp"))

	status := checkWeather(string(PM25))

	log.Println(err)
	switch d {
	case "PSI":
		return string(PSI)
	case "PM25":
		return string(PM25)
	case "Temp":
		return string(Temp)
	default:
		return status
	}
}
