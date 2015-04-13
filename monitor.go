package main

import (
	//"fmt"
	"net/http"
	"strings"
	"strconv"
	"encoding/json"
	"github.com/duncan/db"
	//"github.com/duncan/push"
	"github.com/duncan/scraper"
)

type PushInfo struct {
	UUID       string
	DeviceType string
	Preference int
}

type NearbyCitiesResponse struct {
	Success bool			`json:"success"`
	NearbyCities []db.City	`json:"nearby_cities"`
}

func getJSONStatusMessage(msg string) []byte {
	statusMap := map[string]string{"status": msg}
	jsonByteArray, _ := json.Marshal(statusMap)
	return jsonByteArray
}

//Debug only. Just tells the other end we're alive.
func debug_only(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(getJSONStatusMessage("This is a debug message."))
}

//For clients getting data
func getData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	parameter := r.URL.Query().Get("q")
	comma_index := strings.Index(parameter, ",")

	//Check that the parameter exists and has a comma
	if len(parameter) != 0 && comma_index != -1 {
		//Retrieve lat/lon parameters
		lat_string := parameter[:comma_index]
		lon_string := parameter[comma_index + 1:]
		//Attempt converting into float64
		lat, lat_err := strconv.ParseFloat(lat_string, 64)
		lon, lon_err := strconv.ParseFloat(lon_string, 64)

		//Check if the parameters are actually floats
		if lat_err == nil && lon_err == nil {
			//Check if latitude and longtitude are valid
			if (-90 <= lat && lat <= 90) && (-180 <= lon && lon <= 180) {
				nearbyLocs := db.GetNearbyLocations(lat, lon)
				var root NearbyCitiesResponse;
				if nearbyLocs != nil {
					root = NearbyCitiesResponse{Success : true, NearbyCities : nearbyLocs}
				} else {
					root = NearbyCitiesResponse{Success : false, NearbyCities : make([]db.City, 0)}
				}
				niceJson, _ := json.Marshal(root)
				w.Write(niceJson)
			} else {
				w.Write(getJSONStatusMessage("invalid"))
			}
		} else {
			w.Write(getJSONStatusMessage("invalid"))
		}
	} else {
		w.Write(getJSONStatusMessage("invalid"))
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

	if (k.DeviceType == "iOS") && (len(k.UUID) != 64) {
		w.Write(getJSONStatusMessage("invalid"))
		return
	}

	if (len(k.UUID) == 0) || (len(k.DeviceType) == 0) || (k.Preference == -1) {
		w.Write(getJSONStatusMessage("invalid"))
		return
	}

	saveStatus := db.SavePushDevice(k.UUID, k.DeviceType, k.Preference)

	//Send to database
	if saveStatus {
		w.Write(getJSONStatusMessage("success"))
	} else {
		w.Write(getJSONStatusMessage("failure"))
	}
}

func main() {
	go scraper.ScrapeInterval()

	http.HandleFunc("/", debug_only)
    http.HandleFunc("/api/nearby", getData)
    http.HandleFunc("/post", handlePushDevice)
    http.ListenAndServe(":5000", nil)
}