package model

import (
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/jurvis/config"
	"github.com/jurvis/push"
	"github.com/jurvis/scrape"
	"github.com/steveyen/gkvlite"
	"log"
	"os"
	"strconv"
	"time"
)

func callPush(pm25 string) {
	int_pm25, err := strconv.Atoi(pm25)
	if err != nil {
		log.Println("unable to convert string")
	}
	if int_pm25 > 100 {
		var status string
		if int_pm25 > 200 {
			status = "The air is now hazardous, avoid the outdoors!"
		} else {
			status = "The air is now in an unhealthy range, take care."
		}
		push.PushNotif(status)
	} else {
		log.Println("PM2.5 less than 100, nothing to push.")
	}
}

func HandlePush() {
	log.Println("Spawning APNS...")

	log.Println("Fetching Data for Push...")
	f4, err := os.Open("/tmp/weather.gkvlite")
	s4, err := gkvlite.NewStore(f4)
	c4 := s4.GetCollection("weatherData")
	defer s4.Close()
	s4.Flush()

	PM25, err := c4.Get([]byte("PM25"))
	callPush(string(PM25))
	if err != nil {
		log.Println(err)
	}

	// set up a goroutine to run APNS
	ticker := time.NewTicker(1 * time.Hour)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("Fetching Data for Push...")
				f3, err := os.Open("/tmp/weather.gkvlite")
				s3, err := gkvlite.NewStore(f3)
				c3 := s3.GetCollection("weatherData")
				defer s3.Close()
				s3.Flush()

				PM25, err := c3.Get([]byte("PM25"))
				callPush(string(PM25))
				if err != nil {
					log.Println(err)
				}
			case <-quit:
				ticker.Stop()
				log.Println("Stopped the ticker!")
				return
			}
		}
	}()
}

func tweetData(pm25 string, psi string) {
	cfg := config.TwitterConfig()

	anaconda.SetConsumerKey(cfg.Application.ApiKey)
	anaconda.SetConsumerSecret(cfg.Application.ApiSecret)

	api := anaconda.NewTwitterApi(cfg.Consumer.Token, cfg.Consumer.Secret)
	advisory := checkWeather(pm25)
	s := fmt.Sprintf("'%s' Current PSI: %s, PM2.5: %s. #sghaze", advisory, psi, pm25)
	_, err := api.PostTweet(s, nil)
	if err != nil {
		log.Println(err)
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
