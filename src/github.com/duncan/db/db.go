package db

import (
	"fmt"
	"strconv"
	"log"
	"strings"
	"github.com/duncan/config"
	//"gopkg.in/redis.v2"
	"database/sql"
	_ "github.com/lib/pq"
	"errors"
)

type CountrySource struct {
	Url string
}

type City struct {
	Id int				`json:"id"`
	Name string			`json:"city_name"`
	AdvisoryCode int	`json:"advisory_code"`
	Data int			`json:"data"`
	Temp int			`json:"temperature"`
	ScrapeTime int64 	`json:"scrapetime"`
}

//Structs below are for LEGACY clients
type LegacyWeather struct {
	PM25 string		`json:"PM25"`
	PSI string 		`json:PSI"`
	Temp string 	`json:Temp"`
}

type LegacyCity struct {
	Status string			`json:"status"`
	Weather LegacyWeather	`json:"weather"`
}

var db_config config.DbCfg = config.DbConfig()

//var Redis_URLs *redis.Client = redis.NewClient(&redis.Options{Network:"tcp", Addr:db_config.Redis.Address, DB:0});
//var Redis_DataHolder *redis.Client = redis.NewClient(&redis.Options{Network:"tcp", Addr:db_config.Redis.Address, DB:1});

var PQ_USER, PQ_PASS, PQ_DBNAME, PQ_SSLMODE string = db_config.Database.Username, db_config.Database.Password, db_config.Database.Dbname, "disable"

//Save data into DB
func SaveData(cities []City) {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", PQ_USER, PQ_PASS, PQ_DBNAME, PQ_SSLMODE))

	if err != nil {
		log.Fatal(err)
		return
	}

	defer db.Close()

	query := "BEGIN; CREATE TEMPORARY TABLE newvals (id INTEGER, city_name VARCHAR(100), data INTEGER, temp INTEGER, advisory INTEGER, timestamp BIGINT); "
	query += "INSERT INTO newvals (id, city_name, data, temp, advisory, timestamp) VALUES "

	for _, city := range cities {
		//Overwrite existing Id if it exists.
		older_entry, er := GetSavedData(city.Id)
		id := city.Id
		if er == nil {
			if city.Data == -1 {
				//Use older values
				city.Data = older_entry.Data
				city.AdvisoryCode = older_entry.AdvisoryCode
				city.ScrapeTime = older_entry.ScrapeTime
			}
		} else if UseNextAvailableId() {
			id = GetNextAvailableId()
		}
		query += fmt.Sprintf("(%d, '%s', %d, %d, %d, %d), ", id, city.Name, city.Data, city.Temp, city.AdvisoryCode, city.ScrapeTime)
	}

	query = query[:len(query) - 2] //Remove the last ', '
	query += "; LOCK TABLE data IN EXCLUSIVE MODE; "
	query += "UPDATE data SET id = newvals.id, city_name = newvals.city_name, data = newvals.data, temp = newvals.temp, advisory = newvals.advisory, timestamp = newvals.timestamp FROM newvals WHERE newvals.id = data.id; "

	query += "INSERT INTO data SELECT newvals.id, newvals.city_name, newvals.data, newvals.temp, newvals.advisory, newvals.timestamp FROM newvals LEFT OUTER JOIN data ON (data.id = newvals.id) WHERE data.id IS NULL; "
	query += "COMMIT;"

	_, error := db.Exec(query)

	if error != nil {
		log.Fatal(error)
		return
	}
}

//Return legacy status string. Not present in Nebulo 2.0+
func getLegacyStatus(pm25 int) string {
	if pm25 > 200 {
		return "Stay Indoors."
	} else if pm25 > 100 {
		return "The Air Is Bad."
	} else if pm25 > 50 {
		return "Moderate."
	} else {
		return "It's Clear."
	}
}

//Return Central, Singapore's data for legacy calls
func GetLegacyData() *LegacyCity {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", PQ_USER, PQ_PASS, PQ_DBNAME, PQ_SSLMODE))

	if err != nil {
		log.Fatal(err)
		return nil
	}

	defer db.Close()

	query, er := db.Query("SELECT AVG(data) AS \"data\", AVG(temp) AS \"temp\" FROM data WHERE city_name LIKE '%Singapore';")

	if er != nil {
		return nil
	}

	defer query.Close()

	for query.Next() {
		var data, temp float64
		query.Scan(&data, &temp)
		return &LegacyCity{Status: getLegacyStatus(int(data)), Weather: LegacyWeather{PM25: strconv.Itoa(int(data)), PSI: "N/A", Temp: strconv.Itoa(int(temp))}}
	}

	return nil
}

//Check whether need next avail id
func UseNextAvailableId() bool {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", PQ_USER, PQ_PASS, PQ_DBNAME, PQ_SSLMODE))

	if err != nil {
		log.Fatal(err)
	}

	locations, er := db.Query("SELECT COUNT(*) AS count FROM data")

	defer locations.Close()

	if er != nil {
		log.Fatal(er)
	}

	var count int 

	for locations.Next() {
		locations.Scan(&count)
	}

	defer db.Close()

	if count == 0 {
		return false
	} else {
		return true
	}
}

//Return next available index for country (for appending to bottom)
func GetNextAvailableId() int {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", PQ_USER, PQ_PASS, PQ_DBNAME, PQ_SSLMODE))

	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	locations, er := db.Query("SELECT COUNT(*) AS count FROM data;")

	defer locations.Close()

	if er != nil {
		log.Fatal(er)
	}

	for locations.Next() {
		var count int
		locations.Scan(&count)
		return count
	}

	return -1
}

//Get the saved data for comparison
func GetSavedData(id int) (City, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", PQ_USER, PQ_PASS, PQ_DBNAME, PQ_SSLMODE))

	if err != nil {
		log.Fatal(err)
		return City{}, errors.New("Error accessing db!")
	}

	defer db.Close()

	locations, er := db.Query(`SELECT * FROM data WHERE id=$1`, id)

	defer locations.Close()

	if er != nil {
		log.Fatal(er)
		return City{}, errors.New("Error running db query!")
	}

	for locations.Next() {
		var id int
		var city_name string
		var data int
		var temp int
		var advisory int
		var scrapetime int64
		locations.Scan(&id, &city_name, &data, &temp, &advisory, &scrapetime)
		return City{id, city_name, advisory, data, temp, scrapetime}, nil
	}

	return City{}, errors.New("Nothing to return!")
}

//Return all cities
func GetAllLocations() []City {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", PQ_USER, PQ_PASS, PQ_DBNAME, PQ_SSLMODE))

	if err != nil {
		log.Fatal(err)
		return nil
	}

	defer db.Close()

	var cities []City
	//This query automatically ignores data values of -1.
	locations, er := db.Query(`SELECT * FROM data ORDER BY id ASC;`)

	defer locations.Close()

	if er != nil {
		log.Fatal(er)
		return nil
	}

	for locations.Next() {
		var id int
		var city_name string
		var data int
		var temp int
		var advisory int
		var scrapetime int64

		locations.Scan(&id, &city_name, &data, &temp, &advisory, &scrapetime)
		cities = append(cities, City{id, city_name, advisory, data, temp, scrapetime})
	}
	return cities
}

//Return the closest locations based on lat and lon. Uses PostgreSQL extensions
func GetNearbyLocations(lat, lng float64) []City {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", PQ_USER, PQ_PASS, PQ_DBNAME, PQ_SSLMODE))

	if err != nil {
		log.Fatal(err)
		return nil
	}

	defer db.Close()

	var cities []City
	//This query automatically ignores data values of -1.
	locations, er := db.Query(`SELECT locations.city_name, earth_distance(ll_to_earth($1, $2), ll_to_earth(lat, lng)) as distance FROM locations INNER JOIN data ON data.city_name = locations.city_name WHERE (EXISTS (SELECT 1 FROM data WHERE data.city_name = locations.city_name)) AND (data.data != -1) AND (data.temp != 99999) ORDER BY distance ASC LIMIT 5;`, lat, lng)

	defer locations.Close()

	if er != nil {
		log.Fatal(er)
		return nil
	}

	for locations.Next() {
		var id int
		var distance float64
		var city_name string
		var data int
		var temp int
		var advisory int
		var scrapetime int64
		locations.Scan(&city_name, &distance)
		d, errr := db.Query(`SELECT * FROM data WHERE city_name=$1;`, city_name)

		if errr != nil {
			log.Fatal(errr)
			return nil
		}

		d.Next()
		d.Scan(&id, &city_name, &data, &temp, &advisory, &scrapetime)
		d.Close()
		cities = append(cities, City{id, city_name, advisory, data, temp, scrapetime})
	}
	return cities
}

//Remove a device from push database
func RemovePushDevice(uuid, deviceType string) bool {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", PQ_USER, PQ_PASS, PQ_DBNAME, PQ_SSLMODE))

	if err != nil {
		log.Fatal(err)
		return false
	}

	defer db.Close()

	var table_name string

	if strings.EqualFold(deviceType, "Android") {
		table_name = "push_android"
	} else if strings.EqualFold(deviceType, "iOS") {
		table_name = "push_ios"
	} else {
		//What is this rogue OS?
		log.Printf("Captured a rogue device push with unidentified OS: %s. Request denied.\n", deviceType)
		return false
	}

	log.Printf("Removing %s device with UUID: %s\n", deviceType, uuid)

	result, error := db.Exec("DELETE FROM $1 WHERE uuid='$2';", table_name, uuid) //Default to NULL for id

	if error != nil {
		log.Println("Error occurred removing push device")
		return false
	}

	ra, er := result.RowsAffected()
	if er != nil {
		return false
	}

	log.Printf("Removed %s Push Device. Rows Affected: %d\n", deviceType, ra)
	if ra == 0 {
		return false
	}
	return true
}

//Save a push device into DB
func SavePushDevice(uuid, deviceType string, preference int) bool{
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", PQ_USER, PQ_PASS, PQ_DBNAME, PQ_SSLMODE))

	if err != nil {
		log.Fatal(err)
		return false
	}

	defer db.Close()

	var table_name string

	if strings.EqualFold(deviceType, "Android") {
		table_name = "push_android"
	} else if strings.EqualFold(deviceType, "iOS") {
		table_name = "push_ios"
	} else {
		//What is this rogue OS?
		log.Printf("Captured a rogue device push with unidentified OS: %s. Request denied.\n", deviceType)
		return false
	}

	log.Printf("Saving %s device with UUID: %s\n", deviceType, uuid)

	result, error := db.Exec(fmt.Sprintf("INSERT INTO %s (uuid, id) VALUES ('%s', %d);", table_name, uuid, preference)) //Default to NULL for id

	if error != nil {
		log.Println("Error occurred saving push device")
		return false
	}

	ra, er := result.RowsAffected()
	if er != nil {
		return false
	}

	log.Printf("Saved %s Push Device. Rows Affected: %d\n", deviceType, ra)
	if ra == 0 {
		return false
	}
	return true
}

//Helper method to get devices with a certain preference from a certain table
func GetDevicesByPreference(preference int, table_name string) []string {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", PQ_USER, PQ_PASS, PQ_DBNAME, PQ_SSLMODE))

	if err != nil {
		log.Fatal(err)
		return []string{}
	}

	defer db.Close()

	var devices []string

	rows, er := db.Query(fmt.Sprintf("SELECT uuid FROM %s WHERE id=$1", table_name), preference)

	defer rows.Close()

	if er != nil {
		log.Fatal(er)
	}

	for rows.Next() {
		var uuid string
		err = rows.Scan(&uuid)
		devices = append(devices, uuid)
	}

	return devices
}

//Get Android devices with a certain preference
func GetAndroidDevicesByPreference(preference int) []string {
	return GetDevicesByPreference(preference, "push_android")
}

//Get iOS devices with a certain preference
func GetiOSDevicesByPreference(preference int) []string {
	return GetDevicesByPreference(preference, "push_ios")
}