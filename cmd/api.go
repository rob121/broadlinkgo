package main

import (
	"net/http"
	"strconv"
	"strings"
	"github.com/gorilla/mux"
	"fmt"
)

func apiStatusHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-type", "text/plain")

	var payload = make(map[string]interface{})

	payload["commands"] = getCommands()

	payload["macros"] = getMacros()

	ct := len(Devices)

	payload["devices_found"] = ct

	payload["devices"] = Devices

	payload["equipment"] =  getEquipments()

	payload["types"] = deviceTypes()

	respond(noCache(w), 200, strconv.Itoa(ct)+" Devices found", payload)

}

func apiMacroHandler(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path

	device := r.Form.Get("device")

	path = strings.Replace(path, " ", "_", -1)
	path = strings.ToLower(path)

	parts := strings.Split(strings.Replace(path, "/macro/", "", -1), "/")


	status := false

	var state = make(map[string]bool)

	var output = make(map[string]interface{})

	for _, v := range parts {

		if strings.Contains(v, ":") {

			cmdset := strings.Split(v, ":")
			rep, _ := strconv.Atoi(cmdset[1])



			dev,e := getDevice(device)

			if(e!=nil){
				state[cmdset[0]] = false
				continue
			}

			status = dev.Execute(cmdset[0], rep)

			state[cmdset[0]] = status

		} else {


			dev,e := getDevice(device)

			if(e!=nil){

				continue
			}

			status = dev.Execute(v, 1)



			state[v] = status
		}

	}

	output["commands"] = state

	respond(noCache(w), 200, "Macro executed", output)

}

func apiRemotesHandler(w http.ResponseWriter, r *http.Request) {

  data := make(map[string]interface{})

  rem := getRemotes()

  for kk,re := range rem {

	  for k, b := range re.Buttons {

		  c, e := getCommand(b.Command)

		  if (e == nil) {
			  rem[kk].Buttons[k].Label = c.Label
			  rem[kk].Buttons[k].Command = fmt.Sprintf("/device/%s/cmd/%s", re.Device, c.ToCmd())

		  } else {

			  m, e := getMacro(b.Command)

			  if (e == nil) {
				  rem[kk].Buttons[k].Label = m.Label
				  rem[kk].Buttons[k].Command = fmt.Sprintf("/device/%s/macro/%s", re.Device, m.ToCmd())

			  }
		  }

	  }
  }



	data["remotes"] = rem


  respond(w, 200, "OK", data)

}

func apiDeviceHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	device := vars["device"]

	typ := vars["type"]

	r.URL.Path = strings.Replace(r.URL.Path, fmt.Sprintf("%s/%s","/device",device), "", -1)

	r.ParseForm()

	r.Form.Set("device", device)

	switch(typ){
	case "cmd":
		apiCmdHandler(noCache(w), r)
	case "macro":
		apiMacroHandler(noCache(w),r)
	}

}

func apiCmdHandler(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path

	path = strings.Replace(path, " ", "_", -1)

	path = strings.ToLower(path)

	device := r.Form.Get("device")

	parts := strings.Split(path, "/")

	cmd := ""

	status := false

	if parts[2] != "" {

		fn := parts[2]

		dev,derr:= getDevice(device)

		if(derr!=nil){

			respond(w, 500, "Command NOT Executed "+derr.Error(), "")
			return
		}

		if strings.Contains(fn, ":") {

			cmdset := strings.Split(fn, ":")
			rep, _ := strconv.Atoi(cmdset[1])
			cmd = cmdset[0]

			status = dev.Execute(cmd, rep)


		} else {

			cmd = fn

			status = dev.Execute(cmd, 1)

		}

		if status == true {

			respond(w, 200, "Command "+cmd+" executed", "")
			return
		}

	}
	respond(w, 500, "Command NOT Executed", "")

	return

}