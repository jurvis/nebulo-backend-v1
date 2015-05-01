package weather

import (
	"encoding/json"
	"net/http"
	"fmt"
	"strings"
	"io/ioutil"
	"log"
	"github.com/duncan/db"
	"strconv"
)

var client *http.Client = &http.Client{}
var INVALID_TEMP float64 = 99999

var failsafeCache map[string]WeatherData = make(map[string]WeatherData)

type OWMWeatherResponse struct {
	Main struct {
		Temp float64		`json:"temp"`
		Pressure float64	`json:"pressure"`
		Humidity float64	`json:"humidity"`
	} `json:"main"`
	Name string 			`json:"name"`
}

type YahooWeatherResponse struct {
	Query struct {
		Results struct {
			Weather struct {
				Rss struct {
					Channel struct {
						Item struct {
							Title string `json:"title"`
							Condition struct {
								Temp string `json:"temp"`
							} `json:"condition"`
						} `json:"item"`
					} `json:"channel"`
				} `json:"rss"`
			} `json:"weather"`
		} `json:"results"`
	} `json:"query"`
}

type WeatherData struct {
	Temp float64
}

func ClearCache() {
	for k := range failsafeCache {
		delete(failsafeCache, k)
	}
}

//Get weather data
func GetWeather(id, name, failsafe string) WeatherData {
	ClearCache()
	owm := GetOWMWeather(name, failsafe)
	if owm.Temp == INVALID_TEMP {
		ClearCache()
		//Use Yahoo
		yahoo := GetYahooWeather(name, failsafe)
		if yahoo.Temp == INVALID_TEMP {
			//Really no way
			older_entry, er := db.GetSavedData(id)
			if er == nil {
				return WeatherData{Temp: float64(older_entry.Temp)}
			} else {
				return WeatherData{Temp: INVALID_TEMP}
			}
		} else {
			return yahoo
		}
	} else {
		return owm
	}	
}

//Get Yahoo Weather
func GetYahooWeather(name, failsafe string) WeatherData {
	if len(name) == 0 {
		return WeatherData{Temp: INVALID_TEMP}
	}
	//Re-use cache
	cached, exists := failsafeCache[name]
	if exists {
		return cached
	}
	url := "https://query.yahooapis.com/v1/public/yql?q=SELECT%20*%20FROM%20weather.bylocation%20WHERE%20location%3D'" + strings.Replace(name, " ", "", -1) + "'%20AND%20unit%3D%22c%22&format=json&env=store%3A%2F%2Fdatatables.org%2Falltableswithkeys&callback="
	req, _ := http.NewRequest("GET", url, nil)
	
	//req.Header.Set("User-Agent", "")
	req.Header.Set("Connection", "close")
	res, er := client.Do(req)

	if er == nil {
		contents, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
			return GetYahooWeather(failsafe, "")
		}
		var response YahooWeatherResponse
		json.Unmarshal(contents, &response)

		if strings.Contains(response.Query.Results.Weather.Rss.Channel.Item.Title, "not found") {
			return GetYahooWeather(failsafe, "")
		} else {
			temp_string := response.Query.Results.Weather.Rss.Channel.Item.Condition.Temp
			parsed, err := strconv.ParseFloat(temp_string, 64)
			if err == nil {
				good := WeatherData{Temp: parsed}
				failsafeCache[name] = good
				return good
			} else {
				return GetYahooWeather(failsafe, "")
			}
		}
	} else {
		return GetYahooWeather(failsafe, "")
	}
}

//Get weather data, with a failsafe in case the city name is not listed in OpenWeatherMap.
func GetOWMWeather(name, failsafe string) WeatherData {
	if len(name) == 0 {
		return WeatherData{Temp: INVALID_TEMP}
	}
	//Re-use cache
	cached, exists := failsafeCache[name]
	if exists {
		return cached
	}
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s", strings.Replace(name, " ", "%20", -1)), nil)

	//req.Header.Set("User-Agent", "")
	req.Header.Set("Connection", "close")
	res, er := client.Do(req)
	
	if er == nil {
		contents, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
			return GetOWMWeather(failsafe, "")
		}

		var results OWMWeatherResponse
		json.Unmarshal(contents, &results)
		if len(results.Name) > 0 {
			results.Main.Temp -= 273.15 //Convert Kelvin to Celsius
			res.Body.Close()
			good := WeatherData{Temp: results.Main.Temp}
			failsafeCache[name] = good
			return good
		} else {
			return GetOWMWeather(failsafe, "")
		}
	} else {
		return GetOWMWeather(failsafe, "")
	}
}