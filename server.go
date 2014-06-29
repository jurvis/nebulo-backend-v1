package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jurvis/model"
)

func viewData(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]string)
	m["PSI"] = model.RetrieveData("PSI")
	m["PM25"] = model.RetrieveData("PM25")
	m["Temp"] = model.RetrieveData("Temp")

	type airquality struct {
		Status      string            `json:"status"`
		WeatherData map[string]string `json:"weather"`
	}
	group := airquality{
		Status:      model.RetrieveData("status"),
		WeatherData: m,
	}

	log.Println(m)

	b, err := json.Marshal(group)
	if err != nil {
		log.Println("error:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func main() {
	http.HandleFunc("/get", viewData)

	log.Println("Listening on http://localhost:5000/")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
