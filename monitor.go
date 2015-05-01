package main

import (
	"net/http"
	"strconv"
	"encoding/json"
	"github.com/duncan/db"
	"github.com/duncan/scraper"
	"os"
	"log"
	"fmt"
)

type PushInfo struct {
	UUID		string
	DeviceType	string
	Preference	string
	Push		bool
}

type NearbyCitiesResponse struct {
	Success bool			`json:"success"`
	NearbyCities []db.City	`json:"nearby_cities"`
}

type Point struct {
	Lat float64
	Lng float64
	Wait chan []db.City
}

type LegacyJob struct {
	Wait chan db.LegacyCity
}

var jobs chan *Point
var legacy_jobs chan *LegacyJob

func getJSONStatusMessage(msg string) []byte {
	statusMap := map[string]string{"status": msg}
	jsonByteArray, _ := json.Marshal(statusMap)
	return jsonByteArray
}

//Legacy response for older clients, telling them to update
func legacy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	job := LegacyJob{Wait: make(chan db.LegacyCity)}
	legacy_jobs <- &job

	var l db.LegacyCity
	l = <- job.Wait
	json, err := json.Marshal(l)
	if err != nil {
		w.Write(getJSONStatusMessage("failure"))
	} else {
		w.Write(json)
	}
}

//Debug only. Just tells the other end we're alive.
func debug_only(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(getJSONStatusMessage("invalid"))
}

//For clients getting data
func getData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	lat_string := r.URL.Query().Get("lat")
	lng_string := r.URL.Query().Get("lon")

	//Check that required params exist
	if (len(lat_string) != 0) && (len(lng_string) != 0) {
		//Attempt converting into float64
		lat, lat_err := strconv.ParseFloat(lat_string, 64)
		lng, lng_err := strconv.ParseFloat(lng_string, 64)

		//Check if the parameters are actually floats
		if lat_err == nil && lng_err == nil {
			//Check if latitude and lnggtitude are valid
			if (-90 <= lat && lat <= 90) && (-180 <= lng && lng <= 180) {
				p := Point{Lat: lat, Lng: lng, Wait: make(chan []db.City)}
				jobs <- &p
				var nearbyLocs []db.City
				nearbyLocs = <- p.Wait //Wait for db to process and feed back data
				close(p.Wait)
				var root NearbyCitiesResponse;
				if len(nearbyLocs) != 0 {
					root = NearbyCitiesResponse{Success : true, NearbyCities : nearbyLocs}
				} else {
					root = NearbyCitiesResponse{Success : false, NearbyCities : make([]db.City, 0)}
				}
				niceJson, _ := json.Marshal(root)
				w.Write(niceJson)
				return
			} else {
				w.Write(getJSONStatusMessage("invalid"))
				return
			}
		} else {
			w.Write(getJSONStatusMessage("invalid"))
			return
		}
	} else {
		w.Write(getJSONStatusMessage("invalid"))
		return
	}
}

//For clients registering push to backend
func handlePushDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dec := json.NewDecoder(r.Body)
	var k PushInfo
	err := dec.Decode(&k)
	if err != nil {
		w.Write(getJSONStatusMessage("invalid"))
		return
	}

	//Validation of iOS UUID
	if (k.DeviceType == "iOS") && (len(k.UUID) != 64) {
		w.Write(getJSONStatusMessage("invalid"))
		return
	}

	//Validation of valid data
	if (len(k.UUID) == 0) || (len(k.DeviceType) == 0) {
		w.Write(getJSONStatusMessage("invalid"))
		return
	}

	//Validation of preference
	if (k.Push) && (len(k.Preference) == 0) {
		w.Write(getJSONStatusMessage("invalid"))
		return
	}

	var saveStatus bool

	//Depending on what action to execute
	if k.Push {
		saveStatus = db.SavePushDevice(k.UUID, k.DeviceType, k.Preference)
	} else {
		saveStatus = db.RemovePushDevice(k.UUID, k.DeviceType)
	}

	//Send to database
	if saveStatus {
		w.Write(getJSONStatusMessage("success"))
	} else {
		w.Write(getJSONStatusMessage("failure"))
	}
}

func legacy_worker() {
	for j := range legacy_jobs {
		result := *db.GetLegacyData()
		j.Wait <- result
	}
}

func worker() {
	for j := range jobs {
		result := db.GetNearbyLocations(j.Lat, j.Lng)
		j.Wait <- result
	}
}

func main() {
	//Log output
	pwd, _ := os.Getwd()
	f, err := os.OpenFile(fmt.Sprintf(pwd + "/" + "server.log"), os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Println("Cannot open log file for writing! Logs will print to console instead.")
	} else {
		log.SetOutput(f)
	}

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	defer f.Close()

	//Worker Pool
	jobs = make(chan *Point, 100)
	legacy_jobs = make(chan *LegacyJob, 100)

	for w := 1; w <= 20; w++ {
		go worker()
	}
	for w := 1; w <= 20; w++ {
		go legacy_worker()
	}

	fmt.Println("Nebulo Backend starting...")
	log.Println("Backend started")
	fmt.Println("If the server exits with obscure codes, check server.log")

	go scraper.ScrapeInterval()
	http.HandleFunc("/", debug_only)
	http.HandleFunc("/api/nearby", getData)
	http.HandleFunc("/get", legacy)
	http.HandleFunc("/post", handlePushDevice)
	http.ListenAndServe(":5000", nil)

	close(jobs)
}