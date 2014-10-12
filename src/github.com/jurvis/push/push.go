package push

import (
	"github.com/alexjlockwood/gcm"
	apns "github.com/anachronistic/apns"
	"github.com/jurvis/db"
	"log"
)

func PushNotif(alert string) {
	log.Println("Push Notif Started")
	go PushAPNS(alert)
	go PushGCM(alert)
}

func PushGCM(alert string) {
	log.Println("Push GCM Started")
	db, err := db.Dbconnect()
	if err != nil {
		log.Println("Unable to connect to DB")
	}
	defer db.Close()
	var regIDs []string

	rows, err := db.Query("SELECT uuid FROM devicetokens WHERE devicetype='Android'")
	if err != nil {
		log.Println("Unable to run SQL Query")
	}
	for rows.Next() {
		var uuid string
		err = rows.Scan(&uuid)
		regIDs = append(regIDs, uuid)
	}

	// Create the message to be sent.
	data := map[string]interface{}{"message": alert}
	msg := gcm.NewMessage(data, regIDs...)

	// Create a Sender to send the message.
	sender := &gcm.Sender{ApiKey: "AIzaSyC77YPwT-I5QMbYTeYkywSfSz7-ucvGl0Y"}

	// Send the message and receive the response after at most two retries.
	response, err := sender.Send(msg, 2)
	if err != nil {
		log.Println("Failed to send message:", err)
		return
	}
	log.Println(response.Results)
}

func PushAPNS(alert string) {
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

	client := apns.NewClient("gateway.push.apple.com:2195", "NebuloCert.pem", "apns-dev.pem")

	payload := apns.NewPayload()
	payload.Alert = alert
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
