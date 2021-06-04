package main

import(
	"log"
	"time"
	"strconv"

)

func broadlinkDeviceWatch(){

	ticker := time.NewTicker(5 * time.Second)

	ctr := 0

	for range ticker.C {

		err := broadlink.Discover()
		if err != nil {
			log.Println(err)
			continue
		}

		log.Println("Found " + strconv.Itoa(broadlink.Count()) + " devices")

		if broadlink.Count() < 1 {

			log.Println("No devices found")

			ctr++

			if(ctr>10){

				ctr=0

				addDevicesManually()
			}


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

func addDevicesManually(){

	log.Println("Devices not found, adding manually")

  for _,dev := range getDevices(true) {

	  state := broadlink.AddManualDevice(dev.Ip, dev.Id, dev.Type)
      if(state!=nil){

      	log.Println(state)

	  }
  }


}

