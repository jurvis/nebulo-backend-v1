package push

import (
	"log"
	"github.com/alexjlockwood/gcm"
	apns "github.com/anachronistic/apns"
	"github.com/duncan/db"
	"github.com/duncan/config"
)

//Should we push alerts?
func ShouldPush(city db.City) bool {
	older_entry, er := db.GetSavedData(city.Id)
	if er == nil {
		if older_entry.Data != city.Data {
			return city.AdvisoryCode >= 3
		}
	}
	return false
}

//Push message
func GetPushAlert(city db.City) string {
	if city.AdvisoryCode == 5 {
		return "The air is now hazardous, avoid the outdoors!"
	} else {
		return "The air is now in an unhealthy range, take care."
	}
}

//Send a push notification (this one calls the other methods)
func Push(city db.City) {
	if ShouldPush(city) {
		go pushAPNS(db.GetiOSDevicesByPreference(city.Id), GetPushAlert(city))
		go pushGCM(db.GetAndroidDevicesByPreference(city.Id), GetPushAlert(city))
	} /*else {
		log.Println("Nothing to push")
	}*/
}

//Push to UUIDs of multiple device types. Separates the UUID into device types and pushes.
func MultiPush(uuids []string, msg string) {
	uuids_ios := make([]string, 0)
	uuids_android := make([]string, 0)

	for i := 0; i < len(uuids); i++ {
		if len(uuids[i]) == 64 {
			uuids_ios = append(uuids_ios, uuids[i])
		} else {
			uuids_android = append(uuids_android, uuids[i])
		}
	}
	go pushAPNS(uuids_ios, msg)
	go pushGCM(uuids_android, msg)
}

//Push to iOS Devices
func pushAPNS(uuids []string, msg string) {
	if len(uuids) == 0 {
		return
	}
	client := apns.NewClient("gateway.push.apple.com:2195", "PushCertificate.pem", "PushKey.pem")

	payload := apns.NewPayload()
	payload.Alert = msg
	payload.Sound = "bingbong.aiff"

	for _, uuid := range uuids {
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

//Push to Android devices
func pushGCM(regIDs []string, msg string) {
	if len(regIDs) == 0 {
		return
	}
	cfg := config.PushConfig()

	// Create the message to be sent.
	data := map[string]interface{}{"message": msg}
	gcm_message := gcm.NewMessage(data, regIDs...)

	// Create a Sender to send the message.
	sender := &gcm.Sender{ApiKey: cfg.GCM.ApiKey}

	// Send the message and receive the response after at most two retries.
	response, err := sender.Send(gcm_message, 2)
	if err != nil {
		log.Println("Failed to send message:", err)
		return
	}
	log.Println(response.Results)
}

//Push to WP
/*
func pushWPNS(msg *string) {}
*/

//Push to Pebble Time
/*
func pushPBTime(msg *string) {}
*/