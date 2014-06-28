package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jurvis/scrape"
)

func viewData(w http.ResponseWriter, r *http.Request) {
	type airquality struct {
		PSI  string
		PM25 string
		Temp string
	}
	group := airquality{
		PSI:  scrape.GetData("PSI"),
		PM25: scrape.GetData("PM25"),
		Temp: scrape.GetData("Temp"),
	}

	b, err := json.Marshal(group)
	if err != nil {
		fmt.Println("error:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func main() {
	http.HandleFunc("/get", viewData)
	http.ListenAndServe(":8080", nil)
}
