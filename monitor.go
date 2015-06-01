package main

import (
	"net/http"
	"strconv"
	"encoding/json"
	"github.com/duncan/db"
	"github.com/duncan/config"
	"github.com/duncan/scraper"
	"github.com/duncan/push"
	"github.com/yvasiyarov/gorelic"
	"os"
	"log"
	"fmt"
	"time"
)

type PushInfo struct {
	UUID		string
	DeviceType	string
	Preference	int
}

type PushList struct {
	Message string
	Devices []string
}

type NearbyCitiesResponse struct {
	Success bool			`json:"success"`
	NearbyCities []NearbyCity	`json:"nearby_cities"`
}

//This is essentially the same as the struct in db.go but it has the timestamp as a formatted string.
type NearbyCity struct {
	Id int				`json:"id"`
	Name string			`json:"city_name"`
	AdvisoryCode int	`json:"advisory_code"`
	Data int			`json:"data"`
	Temp int			`json:"temperature"`
	TimeScraped string 	`json:"time_scraped"`
}

type AllCitiesResponse struct {
	Success bool			`json:"success"`
	Cities []AllCity		`json:"cities"`
}

type AllCity struct {
	Id int 				`json:"id"`
	Name string			`json:"city_name"`
}

type Point struct {
	Lat float64
	Lng float64
	Wait chan []db.City
}

type AllCitiesJob struct {
	Wait chan []db.City
}

type LegacyJob struct {
	Wait chan db.LegacyCity
}

type SavePushJob struct {
	Data PushInfo
	Wait chan bool
}

type AllPushResponse struct {
	Success bool			`json:"success"`
	Devices []string		`json:"devices"`
}

type AllPushJob struct {
	Wait chan []string
}

var jobs chan *Point
var allcities_jobs chan *AllCitiesJob
var legacy_jobs chan *LegacyJob
var savepush_jobs chan *SavePushJob
var allpush_jobs chan *AllPushJob

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

//For all cities
func getAllCities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	acj := AllCitiesJob{Wait: make(chan []db.City)}
	allcities_jobs <- &acj
	var allLocs []db.City
	var formattedResponse []AllCity
	allLocs = <- acj.Wait //Wait for db to process and feed back data
	close(acj.Wait)

	//Import db.City into AllCity
	for _, db_city := range allLocs {
		formattedResponse = append(formattedResponse, AllCity{Id: db_city.Id, Name: db_city.Name})
	}

	var root AllCitiesResponse;
	if len(allLocs) != 0 {
		root = AllCitiesResponse{Success : true, Cities : formattedResponse}
	} else {
		root = AllCitiesResponse{Success : false, Cities : make([]AllCity, 0)}
	}
	niceJson, _ := json.Marshal(root)
	w.Write(niceJson)
	return
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
				var formattedResponse []NearbyCity
				nearbyLocs = <- p.Wait //Wait for db to process and feed back data
				close(p.Wait)

				//Import db.City into NearbyCity
				for _, db_city := range nearbyLocs {
					time_scraped := time.Unix(0, db_city.ScrapeTime * 1000000)
					formattedResponse = append(formattedResponse, NearbyCity{db_city.Id, db_city.Name, db_city.AdvisoryCode, db_city.Data, db_city.Temp, time_scraped.UTC().Format("2006-01-02T15:04:05Z")})
				}

				var root NearbyCitiesResponse
				if len(nearbyLocs) != 0 {
					root = NearbyCitiesResponse{Success : true, NearbyCities : formattedResponse}
				} else {
					root = NearbyCitiesResponse{Success : false, NearbyCities : make([]NearbyCity, 0)}
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

	var saveStatus bool

	job := SavePushJob{Data: k, Wait: make(chan bool)}
	savepush_jobs <- &job
	saveStatus = <- job.Wait

	//Send to database
	if saveStatus {
		w.Write(getJSONStatusMessage("success"))
	} else {
		w.Write(getJSONStatusMessage("failure"))
	}
}

//Get all push devices
func allPushDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	job := AllPushJob{Wait: make(chan []string)}
	allpush_jobs <- &job
	var allPushDevices []string
	allPushDevices = <- job.Wait
	var response AllPushResponse
	if len(allPushDevices) > 0 {
		response = AllPushResponse{Success: true, Devices: allPushDevices}
	} else {
		response = AllPushResponse{Success: true, Devices: make([]string, 0)}
	}
	niceJson, _ := json.Marshal(response)
	w.Write(niceJson)
}

//Push to devices
func pushToDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dec := json.NewDecoder(r.Body)
	var k PushList
	err := dec.Decode(&k)
	if err != nil {
		w.Write(getJSONStatusMessage("invalid"))
		return
	}
	if len(k.Message) == 0 || len(k.Devices) == 0 {
		w.Write(getJSONStatusMessage("invalid"))
		return	
	}
	push.MultiPush(k.Devices, k.Message)
	w.Write(getJSONStatusMessage("success"))
}

//Worker to handle legacy jobs
func legacy_worker() {
	for j := range legacy_jobs {
		result := *db.GetLegacyData()
		j.Wait <- result
	}
}
//Worker to handle all cities jobs
func allcities_worker() {
	for j := range allcities_jobs {
		result := db.GetAllLocations()
		j.Wait <- result
	}
}
//Worker to handle normal jobs
func worker() {
	for j := range jobs {
		result := db.GetNearbyLocations(j.Lat, j.Lng)
		j.Wait <- result
	}
}

//Worker to handle allpush jobs
func allpush_worker() {
	for j := range allpush_jobs {
		result := db.GetAllPushDevices()
		j.Wait <- result
	}
}

//Worker to handle savepush jobs
func savepush_worker() {
	for j := range savepush_jobs {
		result := db.SavePushDevice(j.Data.UUID, j.Data.DeviceType, j.Data.Preference)
		j.Wait <- result
	}
}

//Create the jobs channels to receive jobs
func CreateJobChannels() {
	//Worker Pool
	jobs = make(chan *Point, 100)
	allcities_jobs = make(chan *AllCitiesJob, 500)
	legacy_jobs = make(chan *LegacyJob, 100)
	savepush_jobs = make(chan *SavePushJob, 100)
	allpush_jobs = make(chan *AllPushJob, 100)
}

//Start the workers that will do the jobs
func StartWorkers() {
	for w := 1; w <= 20; w++ {
		go worker()
	}
	for w := 1; w <= 10; w++ {
		go allcities_worker()
	}
	for w := 1; w <= 20; w++ {
		go legacy_worker()
	}
	for w := 1; w <= 10; w++ {
		go savepush_worker()
	}
	for w := 1; w <= 5; w++ {
		go allpush_worker()
	}
}

//Setup the log
func SetupLog() {
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
}

//Code to run when program terminates.
func cleanup() {
	db.CloseDB()
	close(jobs)
	close(allcities_jobs)
	close(legacy_jobs)
}

//Register to clean up when Ctrl+C is pressed
func RegisterSignalCleanup() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(){
		for _ := range c {
			cleanup()
		}
	}()
}

func main() {
	SetupLog()

	CreateJobChannels()
	StartWorkers()

	RegisterSignalCleanup()

	//Initialise db
	db.InitialiseDB()

	fmt.Println("Nebulo Backend starting...")
	log.Println("Backend started")
	fmt.Println("If the server exits with obscure codes, check server.log\n")

	fmt.Println("Starting NewRelic agent...")
	agent := gorelic.NewAgent()
	agent.CollectHTTPStat = true
	agent.Verbose = true
	agent.NewrelicLicense = config.NewRelicConfig().License.Key
	agent.Run()

	go scraper.ScrapeInterval()
	http.HandleFunc("/", agent.WrapHTTPHandlerFunc(debug_only))
	http.HandleFunc("/api/all", agent.WrapHTTPHandlerFunc(getAllCities))
	http.HandleFunc("/api/nearby", agent.WrapHTTPHandlerFunc(getData))
	http.HandleFunc("/internal/push/list", agent.WrapHTTPHandlerFunc(allPushDevices))
	http.HandleFunc("/internal/push/push", agent.WrapHTTPHandlerFunc(pushToDevices))
	http.HandleFunc("/get", agent.WrapHTTPHandlerFunc(legacy))
	http.HandleFunc("/api/post/push", agent.WrapHTTPHandlerFunc(handlePushDevice))
	http.ListenAndServe(":5000", nil)

	cleanup()
}