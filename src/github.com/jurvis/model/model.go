package model

import (
	"encoding/json"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/jurvis/scrape"
	"github.com/steveyen/gkvlite"
	"log"
	"os"
	"strconv"
	"time"
)

type Configuration struct {
	Application []string
	Consumer    []string
}

func tweetData(pm25 string, psi string) {
	c := getTwitterConfig("consumer")
	a := getTwitterConfig("application")
	anaconda.SetConsumerKey(a[0])
	anaconda.SetConsumerSecret(a[1])

	api := anaconda.NewTwitterApi(c[0], c[1])
	advisory := checkWeather(pm25)
	s := fmt.Sprintf("'%s' Current PSI: %s, PM2.5: %s.", advisory, psi, pm25)
	_, err := api.PostTweet(s, nil)
	if err != nil {
		log.Println(err)
	}
}

func getTwitterConfig(kind string) []string {
	file, _ := os.Open("twitter-config.json")
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Println("error: ", err)
	}

	switch kind {
	case "consumer":
		return configuration.Consumer
	case "application":
		return configuration.Application
	default:
		return configuration.Consumer
	}
}

func StoreData() {
	// run this on first run
	file, err := os.Create("/tmp/weather.gkvlite")
	if err != nil {
		log.Println("Unable to create .gkvlite file")
	}
	w := scrape.AQICN_Scrape()
	s, err := gkvlite.NewStore(file)
	defer s.Flush()
	if err != nil {
		log.Println("Cannot create new store")
	}
	c := s.SetCollection("weatherData", nil)
	c.Set([]byte("PSI"), []byte(w.PSI))
	c.Set([]byte("PM25"), []byte(w.PM25))
	c.Set([]byte("Temp"), []byte(w.Temperature))

	tweetData(w.PM25, w.PSI)

	// set up a goroutine to scrape every half Hour
	ticker := time.NewTicker(30 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("Scraping...")
				w := scrape.AQICN_Scrape()
				c2 := s.GetCollection("weatherData")
				c2.Set([]byte("PSI"), []byte(w.PSI))
				c2.Set([]byte("PM25"), []byte(w.PM25))
				c2.Set([]byte("Temp"), []byte(w.Temperature))
				s.Flush()
				tweetData(w.PM25, w.PSI)
			case <-quit:
				ticker.Stop()
				log.Println("Stopped the ticker!")
				return
			}
		}
	}()
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

	f3, err := os.Open("/tmp/weather.gkvlite")
	s3, err := gkvlite.NewStore(f3)
	c3 := s3.GetCollection("weatherData")
	defer s3.Close()
	s3.Flush()

	PSI, err := c3.Get([]byte("PSI"))
	PM25, err := c3.Get([]byte("PM25"))
	Temp, err := c3.Get([]byte("Temp"))
	if err != nil {
		log.Println(err)
	}

	status := checkWeather(string(PM25))

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
