package main

import (
	apns "github.com/anachronistic/apns"
	"github.com/jurvis/db"
	"log"
)

func main() {
	db, err := db.Dbconnect()
	if err != nil {
		log.Println("Unable to connect to DB")
	}
	defer db.Close()
	var u []string

	rows, err := db.Query("SELECT uuid FROM devicetokens WHERE devicetype='iOS'")
	if err != nil {
		log.Println("Unable to run SQL Query")
	}
	for rows.Next() {
		var uuid string
		err = rows.Scan(&uuid)
		u = append(u, uuid)
	}

	client := apns.NewClient("gateway.sandbox.push.apple.com:2195", "NebuloCert.pem", "apns-dev.pem")

	payload := apns.NewPayload()
	payload.Alert = "Hello!"
	payload.Sound = "bingbong.aiff"

	for _, uuid := range u {
		log.Println(uuid)
		pn := apns.NewPushNotification()
		pn.DeviceToken = uuid
		pn.AddPayload(payload)

		resp := client.Send(pn)
		if resp.Error != nil {
			log.Println(resp.Error)
		}
		result, _ := pn.PayloadString()
		log.Println(result)
	}
}
