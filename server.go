package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jurvis/model"
)

func viewData(w http.ResponseWriter, r *http.Request) {
	type airquality struct {
		PSI  string
		PM25 string
		Temp string
	}
	group := airquality{
		PSI:  model.RetrieveData("PSI"),
		PM25: model.RetrieveData("PM25"),
		Temp: model.RetrieveData("Temp"),
	}

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
