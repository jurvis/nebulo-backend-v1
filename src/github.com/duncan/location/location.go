package location

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

type OSMElement struct {
	Elements []Location		`json:"elements"`
}

type Location struct {
	Lat float64	`json:"lat"`
	Lng float64	`json:"lon"`
}

//Get weather data, with a failsafe in case the city name is not listed in OpenWeatherMap.
func GetLocation(name, failsafe string) Location {
	if len(name) == 0 {
		return Location{Lat: INVALID_TEMP, Lng: INVALID_TEMP}
	}
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://overpass-api.de/api/interpreter?data=[out:json];node[name=\"%s\"];out;", strings.Replace(name, " ", "%20", -1)), nil)

	//req.Header.Set("User-Agent", "")
	req.Header.Set("Connection", "close")
	res, er := client.Do(req)
	
	if er == nil {
		contents, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
		}
		var response OSMElement
		//It's returning 0 for cities it can't find. Fix it
		json.Unmarshal(contents, &response)

		if len(response.Elements) == 0{
			return GetWeather(failsafe, "")
		}
		res.Body.Close()
		return response.Elements[0]
	} else {
		return GetWeather(failsafe, "")
	}
}