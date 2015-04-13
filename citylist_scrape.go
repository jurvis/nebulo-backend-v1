package main

import (
	"github.com/PuerkitoBio/goquery"
	"fmt"
	"log"
	"strings"
	"strconv"
)

type CityEntry struct {
	Name string
	Url string
}

var countries []string
var cities []CityEntry
var startingIndex int = 0

func generateSQL() string {
	tempIndex := startingIndex
	statement := "INSERT INTO data (id, city_name) VALUES"
	for _, city := range cities {
		statement += fmt.Sprintf(" (%d, '%s')", tempIndex, city.Name)
		tempIndex++
	}
	statement += ";"
	return statement
}

func generateRedisMass() string {
	tempIndex := startingIndex
	statement := ""
	for _, city := range cities {
		statement += fmt.Sprintf("SET %d %s\n", tempIndex, city.Url)
		tempIndex++
	}
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

func main() {
	doc, err := goquery.NewDocument("http://aqicn.org/city/all")
	if err != nil {
		log.Fatal(err)
	}

	var inputs, validCountries []string
	var input string

	fmt.Println("LIST OF COUNTRIES")
	doc.Find("div.whitebody center div.citytreehdr a").Each(func(i int, s *goquery.Selection) {
		fmt.Printf("%d. %s\n", i, s.Text())
		countries = append(countries, s.Text())
	})

	fmt.Print("Enter your choice (e.g. 1): ")
	fmt.Scanf("%s", &input) //<--- here
	inputs = strings.Split(input, ",")

	for _, input := range inputs {
		integer, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Invalid integer '%s'. Aborted.\n", input)
			return
		}
		validCountries = append(validCountries, countries[integer])
	}

	fmt.Println("Obtaining cities for", validCountries)

	doc.Find("div.whitebody center").Each(func(i int, s *goquery.Selection) {
		s.Find("div[style*=\"max-width:80\\%;font-size:18px;\"]").Each(func(ii int, ss *goquery.Selection) {
			country := s.Find("div[style*=\"width:200px;background-color:#9ebac8;color:white;font-size:21px;padding:10px;margin:10px;margin-top:25px;\"]").Eq(ii).Text()
			if !contains(validCountries, country) {
				return
			}
			//fmt.Println("Country:", country)
			a := ss.Find("a")
			for i := range a.Nodes {
				ddd := a.Eq(i)
				href, _ := ddd.Attr("href")
				cities = append(cities, CityEntry{Name: ddd.Text(), Url: href})
			}
		})
	})

	fmt.Printf("Found %d cities.\n\n", len(cities))

	//Starting index (in case adding more cities later on)
	fmt.Print("Enter starting index: ")
	fmt.Scanf("%d", &startingIndex)

	fmt.Println("\n")
	fmt.Println("SQL Statement:")
	fmt.Println(generateSQL(), "\n")
	fmt.Println("Redis Mass:")
	fmt.Println(generateRedisMass())
}