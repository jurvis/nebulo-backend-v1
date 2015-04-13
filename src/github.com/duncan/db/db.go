package db

import (
	"fmt"
	"strconv"
	"log"
	"strings"
	"github.com/duncan/config"
	"gopkg.in/redis.v2"
	"database/sql"
	_ "github.com/lib/pq"
)

type CityURL struct {
	Id int
	Url string
}

type City struct {
	Id int				`json:"id"`
	City_name string	`json:"city_name"`
	AdvisoryCode int	`json:"advisory_code"`
	Pm25 string			`json:"PM25"`
	Temperature string	`json:"temperature"`
}

var db_config config.DbCfg = config.DbConfig()

var Redis_URLs *redis.Client = redis.NewClient(&redis.Options{Network:"tcp", Addr:db_config.Redis.Address, DB:0});
var Redis_DataHolder *redis.Client = redis.NewClient(&redis.Options{Network:"tcp", Addr:db_config.Redis.Address, DB:1});

var PQ_USER, PQ_PASS, PQ_DBNAME, PQ_SSLMODE string = db_config.Database.Username, db_config.Database.Password, db_config.Database.Dbname, "disable"

//Return ALL locations (id and URL)
func GetAllCityURLs() []CityURL {
	keys, err := Redis_URLs.Keys("*").Result()
	var cities []CityURL

	if err != nil {
		fmt.Println("Error retrieving Redis URL keys")
		return nil
	}

	for _, key := range keys {
		key_int, error := strconv.Atoi(key)
		if error == nil {
			value, er := Redis_URLs.Get(key).Result()
			if er == nil {
				cities = append(cities, CityURL{Id:key_int, Url:value})
			} else {
				fmt.Printf("Error retrieving value of key '%s'\n", key)
			}
		} else {
			fmt.Printf("Error converting string '%s' to int\n", key)
		}
	}

	return cities
}

//Save location result into REDIS for Temp storage
func SaveLocationResult(id, data, temp int) {
	//Redis HMSET (or at least the golang lib) doesn't support hash values being integers.
	id_string := strconv.Itoa(id)
	data_string := strconv.Itoa(data)
	temp_string := strconv.Itoa(temp)
	cmd := Redis_DataHolder.HMSet(id_string, "data", data_string, "temp", temp_string)
	if cmd.Err() != nil {
		fmt.Println("Error occurred saving location result into Redis")
	}
}

//Save REDIS Temp Storage data into SQL
func SaveRedisDataIntoDB() {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", PQ_USER, PQ_PASS, PQ_DBNAME, PQ_SSLMODE))

	if err != nil {
		log.Fatal(err)
		return
	}

	defer db.Close()

	pm25_query := "pm25 = CASE"
	temp_query := "temp = CASE"
	final_query := ""

	keys, err := Redis_DataHolder.Keys("*").Result()

	if err != nil {
		log.Fatal(err)
		//fmt.Println("Error retrieving data held in Redis Redis_DataHolder")
		return
	}

	for _, key := range keys {
		data, errr := Redis_DataHolder.HGet(key, "data").Result()
		temp, errrr := Redis_DataHolder.HGet(key, "temp").Result()
		
		if (errr != nil) || (errrr != nil) {
			continue
		}

		pm25_query += " WHEN id=" + key + " THEN " + data;
		temp_query += " WHEN id=" + key + " THEN " + temp;
	}

	pm25_query += " END"
	temp_query += " END"

	//Run the command
	final_query = "UPDATE data SET " + pm25_query + ", " + temp_query
	_, error := db.Exec(final_query)

	if error != nil {
		log.Fatal(error)
		//fmt.Println("Error occurred transferring from Redis to SQL")
		return
	}

	//ra, _ := result.RowsAffected()
	//fmt.Printf("Transfer from Redis to SQL: %d/%d\n", ra, len(keys))
}

//Wipe Redis' Temp Storage
/*func WipeRedisTempStorage() {

}*/

//Return the closest locations based on lat and lon. Use PostgreSQL.
func GetNearbyLocations(lat, lng float64) []City {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", PQ_USER, PQ_PASS, PQ_DBNAME, PQ_SSLMODE))

	if err != nil {
		log.Fatal(err)
		return nil
	}

	defer db.Close()

	//fmt.Printf("Retrieving locations near lat:%f lon:%f\n", lat, lng)
	var cities []City
	locations, er := db.Query(`SELECT id, earth_distance(ll_to_earth($1, $2), ll_to_earth(lat, lng)) as distance FROM locations ORDER BY distance ASC LIMIT 5;`, lat, lng)

	if er != nil {
		log.Fatal(er)
		return nil
	}

	for locations.Next() {
		var id int
		var distance float64
		var city_name string
		var pm25 int
		var temp int
		locations.Scan(&id, &distance)
		data, errr := db.Query(`SELECT * FROM data WHERE id=$1;`, id)

		if errr != nil {
			log.Fatal(errr)
			return nil
		}

		data.Next()
		data.Scan(&id, &city_name, &pm25, &temp)
		cities = append(cities, City{id, city_name, GetAdvisoryCode(pm25), strconv.Itoa(pm25), strconv.Itoa(temp)})
	}
	return cities
}

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
		fmt.Printf("Captured a rogue device push with unidentified OS: %s. Request denied.\n", deviceType)
		return false
	}

	fmt.Printf("Saving %s device with UUID: %s\n", deviceType, uuid)

	result, error := db.Exec(fmt.Sprintf("INSERT INTO %s (uuid, id) VALUES (%s, %d);", table_name, uuid, preference)) //Default to NULL for id

	if error != nil {
		log.Fatal(error)
		fmt.Println("Error occurred saving push device")
		return false
	}

	ra, er := result.RowsAffected()
	if er != nil {
		return false
	}

	fmt.Printf("Saved %s Push Device. Rows Affected: %d\n", deviceType, ra)
	return true
}

func GetAdvisoryCode(pm25 int) int{
	if pm25 > 200 {
		//status = "Stay Indoors."
		return 3
	} else if pm25 > 100 {
		//status = "The Air Is Bad."
		return 2
	} else if pm25 > 50 {
		//status = "Moderate."
		return 1
	} else {
		//status = "It's Clear."
		return 0
	}
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

	rows, err := db.Query("SELECT uuid FROM $1 WHERE id=$2", table_name, preference)
	if err != nil {
		log.Println("Unable to run SQL Query")
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