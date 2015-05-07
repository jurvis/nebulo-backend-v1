package main

import (
	"github.com/PuerkitoBio/goquery"
	"fmt"
	"log"
	"strings"
	"strconv"
	//"regexp"
	"net/http"
	"net/url"
	"io/ioutil"
	"time"
	"encoding/json"
)

type Country struct {
	Name string
	Cities []CityEntry
}

type CityEntry struct {
	Name string
	Url string
}

type Coordinate struct {
	Lat float64
	Lng float64
}

type AQICNSearchResult struct {
	Key string 	`json:"key"`
}

type AQICNNearbyCity struct {
	G []string 	`json:"g"`
}

type AQICNAndroidResponse struct {
	Nearest []AQICNNearbyCity	`json:"nearest"`
}

type LocationJob struct {
	ArrayIndex int
	City CityEntry
}

type LocationResult struct {
	ArrayIndex int
	StatementFragment string
}

var countries []Country
var languages []string
var startingIndex, coordCount int = 0, 0
var altLanguage string
var client *http.Client = &http.Client{}
var locationArray []string
var locationJobs chan LocationJob = make(chan LocationJob, 3000)
var locationResults chan LocationResult = make(chan LocationResult, 3000)

func locationWorker() {
	for j := range locationJobs {
		coords := getCoordinates(j.City)
		if coords == nil {
			log.Fatal("Could not obtain coordinates for %s, locations generation aborted.\n", j.City.Name)
		}
		locationResults <- LocationResult{ArrayIndex: j.ArrayIndex, StatementFragment: fmt.Sprintf(" ('%s', %f, %f),", j.City.Name, coords.Lat, coords.Lng)}
	}
}

func generateLocationSQL(cities []CityEntry) string {
	statement := "INSERT INTO locations (id, lat, lng) VALUES"
	fmt.Printf("LAT/LON: Processing %2d/%d\r", coordCount, len(cities))
	locationArray = make([]string, len(cities))
	for index, city := range cities {
		locationJobs <- LocationJob{ArrayIndex: index, City: city}
	}
	for i := 0; i < len(cities); i++ {
		result := <- locationResults
		locationArray[result.ArrayIndex] = result.StatementFragment
	}
	for _, complete := range locationArray {
		statement += complete
	}
	sz := len(statement)
	statement = statement[:sz-1]
	statement += ";"
	return statement
}

func contains(s []string, e string) bool {
	for _, a := range s { 
		if strings.EqualFold(a, e) {
			return true
		}
	}
	return false
}

/*func getCoordinates(city CityEntry) *Coordinate {
	req, _ := http.NewRequest("GET", city.Url, nil)
	req.Header.Set("Connection", "close")
	res, er := client.Do(req)
	defer res.Body.Close()

	pattern, err := regexp.Compile(fmt.Sprintf("\"city\":\"%s(?:.*?)\"(?:.*?)\"g\":\\[\"(.*?)\",\"(.*?)\"]}", strings.Replace(city.Name, "/", "\\\\/", -1)))
	if err != nil {
		fmt.Println("Regex pattern compilation error!")
		return nil
	}

	coordCount++
	fmt.Printf("LAT/LON: Processing %d\r", coordCount)

	if er == nil {
		contents, err := ioutil.ReadAll(res.Body)
		if err == nil {
			patternResults := pattern.FindStringSubmatch(string(contents))
			if len(patternResults) != 3 {
				return nil
			}
			lat, lat_err := strconv.ParseFloat(patternResults[1], 64)
			lng, lng_err := strconv.ParseFloat(patternResults[2], 64)
			if (lat_err != nil) || (lng_err != nil) {
				return nil
			}
			return &Coordinate{Lat: lat, Lng: lng}
		} else {
			return nil
		}
		return nil
	} else {
		return nil
	}
}*/

// UrlEncoded encodes a string like Javascript's encodeURIComponent()
func UrlEncoded(str string) (string, error) {
    u, err := url.Parse(str)
    if err != nil {
        return "", err
    }
    return u.String(), nil
}

func searchCity(city CityEntry) string {
	millis := time.Now().UnixNano() / 1000000
	cityNameEncoded, err := UrlEncoded(city.Name)
	if err != nil {
		log.Fatal(err)
	}
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://waqi.aqicn.org/mapi/?lang=en&key&n=28&t=%d&term=%s", millis, cityNameEncoded), nil)
	res, er := client.Do(req)

	if er == nil {
		contents, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		results := make([]AQICNSearchResult, 0)
		json.Unmarshal(contents, &results)
		res.Body.Close()
		return results[0].Key
	} else {
		log.Fatal(er)
	}
	return ""
}

func getCoordinatesWithKey(key string) *Coordinate {
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://aqicn.org/aqicn/json/android/%s", key), nil)
	res, er := client.Do(req)

	if er == nil {
		contents, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		results := &AQICNAndroidResponse{}
		json.Unmarshal(contents, &results)
		firstResult := results.Nearest[0]

		lat, lat_err := strconv.ParseFloat(firstResult.G[0], 64)
		lng, lng_err := strconv.ParseFloat(firstResult.G[1], 64)

		if (lat_err != nil) || (lng_err != nil) {
			log.Fatal("Lat/Lng parsing error: %s\n", firstResult.G)
		}

		res.Body.Close()
		return &Coordinate{lat, lng}
	} else {
		log.Fatal(er)
	}
	return nil
}

func getCoordinates(city CityEntry) *Coordinate {
	cityKey := searchCity(city)
	if len(cityKey) == 0 {
		return nil
	}
	coordCount++
	fmt.Printf("LAT/LON: Processing %2d\r", coordCount)
	return getCoordinatesWithKey(cityKey)
}

//Population of data, required for ALL operations
func populate() {
	doc, err := goquery.NewDocument("http://aqicn.org/city/all")
	if err != nil {
		log.Fatal(err)
	}

	//Populate languages
	/*pattern, err := regexp.Compile("&lang=(.*)")
	if err != nil {
		fmt.Println("Regex pattern compilation error!")
		return
	}*/

	fmt.Println("Retrieving list of languages...")

	/*doc.Find("div#header div#header-in div[style*=\"position:absolute;left:48px;margin-top:5px;font-size:12px;\"]").Each(func(i int, s *goquery.Selection) {
		a := s.Find("a")
		for i := range a.Nodes {
			ddd := a.Eq(i)
			href, _ := ddd.Attr("href")
			languages = append(languages, pattern.FindStringSubmatch(href)[1])
		}
		fmt.Printf("Populated %d languages.\n", len(languages))
	})*/

	//Populate cities under countries
	doc.Find("div.whitebody center").Each(func(i int, s *goquery.Selection) {
		s.Find("div[style*=\"max-width:80\\%;font-size:18px;\"]").Each(func(ii int, ss *goquery.Selection) {
			country := s.Find("div[style*=\"width:200px;background-color:#9ebac8;color:white;font-size:21px;padding:10px;margin:10px;margin-top:25px;\"]").Eq(ii).Text()

			var cities []CityEntry

			a := ss.Find("a")
			for i := range a.Nodes {
				ddd := a.Eq(i)
				href, _ := ddd.Attr("href")
				cities = append(cities, CityEntry{Name: ddd.Text(), Url: href})
			}

			countries = append(countries, Country{Name: country, Cities: cities})
		})
	})
}

func askForDirection() {
	var choice int
	fmt.Print("1 for Mass Insertion, 2 for Language Generator: ")
	fmt.Scanf("%d", &choice)
	if choice == 1 {
		database()
	} else {
		language()
	}
}

func showCountries() []Country {
	var validCountries []Country
	var inputs []string
	var input string

	fmt.Println("LIST OF COUNTRIES")
	/*doc.Find("div.whitebody center div.citytreehdr a").Each(func(i int, s *goquery.Selection) {
		fmt.Printf("%d. %s\n", i, s.Text())
		countries = append(countries, s.Text())
	})*/
	for i := 0; i < (len(countries) / 3); i++ {
		fmt.Printf("%-2d. %-20s\t%-2d. %-20s\t%-2d. %-20s\n", i, countries[i].Name, i + (len(countries) / 3), countries[i + (len(countries)/ 3)].Name, i + 2 * (len(countries) / 3), countries[i + 2 * (len(countries)/ 3)].Name)
	}

	fmt.Print("Enter your choice (e.g. 1) (e.g. 1,9,16): ")
	fmt.Scanf("%s", &input) //<--- here
	inputs = strings.Split(input, ",")

	fmt.Print("Obtaining information for ")

	for _, input := range inputs {
		integer, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Invalid integer '%s'. Aborted.\n", input)
			return nil
		}
		fmt.Print(countries[integer].Name + ", ")
		validCountries = append(validCountries, countries[integer])
	}

	return validCountries
}

//Functions to produce translations
func language() {

}

//Functions to produce database mass insertions
func database() {
	validCountries := showCountries()

	var allCities []CityEntry

	for _, country := range validCountries {
		for _, city := range country.Cities {
			allCities = append(allCities, city)
		}
	}

	fmt.Printf("Total: %d cities\n", len(allCities))

	fmt.Println("\nSQL Statement for locations:")
	fmt.Println(generateLocationSQL(allCities), "\n")
}

func main() {
	//Start location workers
	for i := 0; i < 3; i++ {
		go locationWorker()
	}
	populate()
	askForDirection()
}