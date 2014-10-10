package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jurvis/model"
)

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		log.Printf("Completed in %s", time.Now().Sub(start).String())

		handler.ServeHTTP(w, r)
	})
}

type UUID struct {
	UUID       string
	DeviceType string
}

func postUUID(w http.ResponseWriter, r *http.Request) {
	type result struct {
		Status string
	}

	dec := json.NewDecoder(r.Body)
	var k UUID
	err := dec.Decode(&k)
	if err != nil {
		log.Println(err)
		log.Println(r.Body)
		response := result{status: "Houston, we have a problem."}
		b, err := json.Marshal(response)
		if err != nil {
			log.Println("error:", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
		panic(err)
	}

	response := result{status: "success."}
	b, err := json.Marshal(response)
	if err != nil {
		log.Println("error:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)

	log.Println(k.UUID)
}

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
	model.StoreData()
	http.HandleFunc("/get", viewData)
	http.HandleFunc("/post", postUUID)

	log.Println("Listening on http://localhost:5000/")
	log.Fatal(http.ListenAndServe(":5000", Log(http.DefaultServeMux)))
}
