package weather

import (
	"encoding/json"
	"net/http"
	"fmt"
	"strings"
	"io/ioutil"
	"log"
)

var client *http.Client = &http.Client{}
var INVALID_TEMP float64 = 99999

var failsafeCache map[string]WeatherData = make(map[string]WeatherData)

type WeatherSubData struct {
	Temp float64		`json:"temp"`
	Pressure float64	`json:"pressure"`
	Humidity float64	`json:"humidity"`
}

type WeatherData struct {
	Main WeatherSubData	`json:"main"`
}

type WeatherResponse struct {
	Message string 		`json:"message"`
}

func ClearCache() {
	for k := range failsafeCache {
		delete(failsafeCache, k)
	}
}

//Get weather data, with a failsafe in case the city name is not listed in OpenWeatherMap.
func GetWeather(name, failsafe string) WeatherData {
	if len(name) == 0 {
		return WeatherData{WeatherSubData{Temp: INVALID_TEMP}}
	}
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s", strings.Replace(name, " ", "%20", -1)), nil)

	//req.Header.Set("User-Agent", "")
	req.Header.Set("Connection", "close")
	res, er := client.Do(req)
	
	if er == nil {
		contents, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
		}
		var response WeatherResponse
		json.Unmarshal(contents, &response)

		if response.Message != "" {
			if _, ok := failsafeCache[failsafe]; !ok {
				failsafeCache[failsafe] = GetWeather(failsafe, "")
			}
			sad := failsafeCache[failsafe]
			return sad
		}

		var results WeatherData
		json.Unmarshal(contents, &results)
		results.Main.Temp -= 273.15 //Convert Kelvin to Celsius
		res.Body.Close()
		return results
	} else {
		return GetWeather(failsafe, "")
	}
}