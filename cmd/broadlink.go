package main

import(
	"log"
	"time"
	"strconv"

)

func broadlinkDeviceWatch(){

	ticker := time.NewTicker(5 * time.Second)

	for range ticker.C {

		err := broadlink.Discover()
		if err != nil {
			log.Println(err)
			continue
		}

		log.Println("Found " + strconv.Itoa(broadlink.Count()) + " devices")

		if broadlink.Count() < 1 {

			log.Println("No devices found")

		} else {

			log.Println("Devices Found, updating check interval")

			syncBroadlinkDevices()

			ticker.Stop()
			ticker = time.NewTicker(300 * time.Second) //look every 5

		}

	}

}


func syncBroadlinkDevices(){

	ids := broadlink.DeviceIds()

	for mac,d := range ids {

		AddDevice(mac,"",d[0],d[1])

	}
	//refresh internal devices list

	getDevices()

}

