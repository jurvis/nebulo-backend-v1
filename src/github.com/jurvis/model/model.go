package model

import (
	"fmt"
	"github.com/jurvis/scrape"
	"github.com/steveyen/gkvlite"
	"log"
	"os"
	"time"
)

func storeData() {
	fmt.Println("Scraping...")
	w := scrape.AQICN_Scrape()
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
	log.Println(err)
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
