package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jurvis/model"
	"github.com/yvasiyarov/gorelic"
)

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		log.Printf("Completed in %s", time.Now().Sub(start).String())

		handler.ServeHTTP(w, r)
	})
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
	agent := gorelic.NewAgent()
	agent.Verbose = true
	agent.NewrelicLicense = ""
	agent.Run()
	http.HandleFunc("/get", agent.WrapHTTPHandlerFunc(viewData))

	log.Println("Listening on http://localhost:5000/")
	log.Fatal(http.ListenAndServe(":5000", Log(http.DefaultServeMux)))
}
