package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/GeertJohan/go.rice"

	"github.com/2opremio/broadlinkgo"
)

var broadlink broadlinkgo.Broadlink
var code string
var port int
var cmdpath string
var mode string

func defaultHandler(w http.ResponseWriter, r *http.Request) {

	//fmt.Fprintln(w,"/status for command info, /learn to learn a command, /macro to send a group of commands (slash separated), cmd to send one command")

	ct := broadlink.Count()

	templateBox, err := rice.FindBox("httpassets")
	if err != nil {
		log.Fatal(err)
	}
	// get file contents as string
	templateString, err := templateBox.String("tmpl/index.html")
	if err != nil {
		log.Fatal(err)
	}
	// parse and execute the template
	tmplMessage, err := template.New("message").Parse(templateString)
	if err != nil {
		log.Fatal(err)
	}

	kd := broadlink.DeviceTypes()

	device_sel := "<select class='form-control' name='deviceType' >"

	kk := ""

	for k, v := range kd {

		kk = strconv.Itoa(k)

		device_sel += "<option value='" + kk + "' >" + v + "</option>"

	}

	device_sel += "</select>"

	tmplMessage.Execute(w, map[string]string{"Mode":mode,"DevicesCT": strconv.Itoa(ct), "DeviceList": device_sel})

}

func macroHandler(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path

	device := r.Form.Get("device")

	path = strings.Replace(path, " ", "_", -1)
	path = strings.ToLower(path)

	parts := strings.Split(strings.Replace(path, "/macro/", "", -1), "/")

	log.Println(parts)

	status := false

	var state = make(map[string]bool)

	var output = make(map[string]interface{})

	for _, v := range parts {

		if strings.Contains(v, ":") {

			cmdset := strings.Split(v, ":")
			rep, _ := strconv.Atoi(cmdset[1])

			status = executeCmd(cmdset[0], rep, device)

			state[cmdset[0]] = status

		} else {

			status = executeCmd(v, 1, device)

			state[v] = status
		}

	}

	output["commands"] = state

	respond(w, 200, "Macro executed", output)

}

func executeCmd(cmd string, repeat int, device string) bool {

	//magic command to help macros
	if cmd == "delay" {

		time.Sleep(1 * time.Second)
		return true

	}

	if repeat == 0 {
		repeat = 1
	}
	
	
	fp := filepath.FromSlash(cmdpath + "commands/cmd_" + cmd + ".txt")

	content, err := ioutil.ReadFile(fp)

	if err != nil {
		log.Println(err)
		return false
	}

	code = string(content)

	for i := 0; i < repeat; i++ {

		broadlink.Execute(device, code)
		time.Sleep(5 * time.Millisecond) //introduce a delay here

	}

	return true

}

func cmdHandler(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path

	path = strings.Replace(path, " ", "_", -1)

	path = strings.ToLower(path)

	device := r.Form.Get("device")

	parts := strings.Split(path, "/")

	cmd := ""

	status := false

	if parts[2] != "" {

		fn := parts[2]

		if strings.Contains(fn, ":") {

			cmdset := strings.Split(fn, ":")
			rep, _ := strconv.Atoi(cmdset[1])
			cmd = cmdset[0]
			status = executeCmd(cmd, rep, device)

		} else {

			cmd = fn

			status = executeCmd(cmd, 1, device)
		}

		if status == true {

			respond(w, 200, "Command "+cmd+" executed", "")
			return
		}

	}
	respond(w, 500, "Command NOT Executed", "")

	return

}

func learnHandler(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	
	w.Header().Set("Content-type", "text/html")

	path = strings.Replace(path, " ", "_", -1)
	path = strings.ToLower(path)

	r.ParseForm()

	frame := ""
	
	
    kd := broadlink.DeviceIds()

	device_sel := "<select class='form-control' name='device' >"

	//kk := ""

	for k, v := range kd {

		//kk = strconv.Itoa(k)

		device_sel += "<option value='" + k + "' >" + k+" ("+v[0]+")</option>"

	}

	device_sel += "</select>"

	

	if r.FormValue("cmd") != "" {

		cmd := r.FormValue("cmd")
		
		dev := r.FormValue("device")
		
		rf := r.FormValue("rf")
		
		src := ``

		if(dev!=""){
			src = src + `/device/` + dev
		}

		src = src + `/learnchild/` + cmd

		if(rf!=""){
			src = src + `/rf/`
		}

		frame = `<h4>Learning device [` + cmd + `]</h4><iframe src='` + src + `' width='100%' height='500px' ></iframe>`


	}
	

	templateBox, err := rice.FindBox("httpassets")
	if err != nil {
		log.Fatal(err)
	}
	// get file contents as string
	templateString, err := templateBox.String("tmpl/learn.html")
	if err != nil {
		log.Fatal(err)
	}
	// parse and execute the template
	tmplMessage, err := template.New("message").Parse(templateString)
	if err != nil {
		log.Fatal(err)
	}
	tmplMessage.Execute(w, map[string]string{"Frame": frame,"DeviceList": device_sel})
	


}

func learnChildHandler(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path

	path = strings.Replace(path, " ", "_", -1)
	path = strings.ToLower(path)
	
	
	device := r.Form.Get("device")

	w.Header().Set("Content-type", "text/html")
	fmt.Fprintln(w, "<style> body{margin:0px;padding:30px;background-color:#000;color:#FFF;}</style>")
	fmt.Fprintln(w, "<pre>")

	parts := strings.Split(path, "/")

	// TODO: len(parts) < 3 is no longer a valid check, with the potential addition of /rf/ (could check "Is RF", without specifying a "Device Name/Action")
	if len(parts) < 3 {

		fmt.Fprintln(w, "Provide a button/device name")
		fmt.Fprintln(w, "</pre>")
		return

	}

	rf := strings.Contains(path, "/rf/")

	if (rf) {
		fmt.Fprintln(w, "Waiting for RF remote. IMPORTANT - press on for 1 second and release until learning is finished <blink>....</blink>")
	}else{
		fmt.Fprintln(w, "Waiting for ir remote presses<blink>....</blink>")
	}

	fmt.Fprintln(w, "&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<br>")

	fn := ""

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	} else {
		log.Println("no flush")
	}

	var data string
	var err error

	if (rf) {

		data, err = broadlink.LearnRF(device)
	}else{
		data, err = broadlink.Learn(device)
	}

	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		fmt.Fprintln(w, "</pre>")
		return
	}

	if len(data) == 0 {
		fmt.Fprintln(w, "Error: have not learned code")
		fmt.Fprintln(w, "</pre>")
		return
	}

	//create a file with command

    var fp = ""
	if data != "" {

		fn = parts[2]
		
		fp = filepath.FromSlash(cmdpath + "commands/cmd_" + fn + ".txt")

		f, err := os.Create(fp)

		if err != nil {
			log.Println(err)
			return
		}

		l, err := f.WriteString(data)
		if err != nil {
			log.Println(err)
			f.Close()
			return
		}

		if l < 1 {
		}

	}

	fmt.Fprintln(w, "Code Detected!")
	fmt.Fprintln(w, data)
	fmt.Fprintln(w, "Code Saved to "+fp)
	
	if(device!=""){
		
	fmt.Fprintln(w, "Use /device/"+device+"/cmd/"+fn+" to trigger the command")	
		
	}else{
	
	fmt.Fprintln(w, "Use /cmd/"+fn+" to trigger the command")
	
	}
	fmt.Fprintln(w, "</pre>")

	return

}

func manualDeviceHandler(w http.ResponseWriter, r *http.Request) {

	//path := r.URL.Path

	w.Header().Set("Content-type", "application/json")

	r.ParseForm()

	if r.Method != "POST" {
		respond(w, 500, "Invalid Request - must POST", "")
		return

	}

	ip := r.FormValue("ip")

	mac := r.FormValue("mac")

	if ip == "" || mac == "" {

		respond(w, 500, "Ip, Mac Required", "")
		return
	}

	deviceType, _ := strconv.Atoi(r.Form.Get("deviceType"))

	state := broadlink.AddManualDevice(ip, mac, deviceType)

	if state != nil {

		respond(w, 500, "Add Error "+state.Error(), "")
		return

	}
	
	//save this device
	

	saveDevices(ip,mac,deviceType)


	respond(w, 200, "Device Added Succesfully", "")

}

func removeDevice(mac string){
	
	dev := getDeviceSaved();
	
	delete(dev,mac);
	
	
	broadlink.RemoveDevice(mac)
	
	fp := filepath.FromSlash(cmdpath+"devices.gob")
	
	file, err := os.Create(fp)
   
    if err == nil { 
       
       
    }
        
    encoder := gob.NewEncoder(file)
     
    if err := encoder.Encode(dev); err != nil {
		
	}
	
	file.Close()
	
}

func getDeviceSaved() map[string][]string{
	
		// Create a file for IO
		
	fp := filepath.FromSlash(cmdpath+"devices.gob")	
	
	byt, err := ioutil.ReadFile(fp)
	
	encodeFile := bytes.NewReader(byt)
	
	if err != nil {
	
	}


	
	decoder := gob.NewDecoder(encodeFile)
	
		// Place to decode into
	out := make(map[string][]string)

	// Decode -- We need to pass a pointer otherwise accounts2 isn't modified
	decoder.Decode(&out)
	

	
	return out
	
}

func saveDevices(ip string,mac string,devicetype int){
	
	
	dev := getDeviceSaved();
	 
	//read, merge, save
	sdev := strconv.Itoa(devicetype)
	
	dev[mac]=[]string{ip,sdev}

	fp := filepath.FromSlash(cmdpath+"devices.gob")	 
	 
	file, err := os.Create(fp)
   
    if err == nil { 
       
       
    }
        
    encoder := gob.NewEncoder(file)
     
    if err := encoder.Encode(dev); err != nil {
		
	}
	
	file.Close()
	
	 
}



func deviceHandler(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	path = strings.Replace(path, " ", "_", -1)
	path = strings.ToLower(path)

	w.Header().Set("Content-type", "text/plain")

	parts := strings.Split(path, "/")

	if len(parts) < 2 {

		//fmt.Fprintln(w,"Must provide a device")

	}

	device := parts[2]

	// fmt.Fprintln(w,device)

	ids := broadlink.DeviceIds()

	if _, ok := ids[device]; ok {
		//do something here
		//  fmt.Fprintln(w,"Device Exists")
		//update path set a from value

		// fmt.Fprintln(w,r.URL.Path)

		r.URL.Path = strings.Replace(r.URL.Path, "/device/"+device, "", -1)

		// fmt.Fprintln(w,r.URL.Path)

		r.ParseForm()

		r.Form.Set("device", device)

		if strings.Contains(path, "/cmd/") {

			cmdHandler(w, r)

		} else if strings.Contains(path, "/macro/") {

			macroHandler(w, r)

		}else if strings.Contains(path, "/learnchild/") {
			
			learnChildHandler(w, r)
			
		}

	} else {

		// fmt.Fprintln(w,"Device Does Not Exist")

	}

}

func removeDeviceHandler(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path

	path = strings.Replace(path, " ", "_", -1)

	path = strings.ToLower(path)

	parts := strings.Split(path, "/")

	cmd := ""

	if r.Method != "POST" {
		respond(w, 500, "Invalid Request - must POST", "")
		return

	}

	if parts[2] == "" {

		respond(w, 500, "Invalid Request", "")
		return

	}

	cmd = parts[2]

	removeDevice(cmd)
	
	respond(w, 200, "Command Removed", "")

}


func removeHandler(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path

	path = strings.Replace(path, " ", "_", -1)

	path = strings.ToLower(path)

	parts := strings.Split(path, "/")

	cmd := ""

	if r.Method != "POST" {
		respond(w, 500, "Invalid Request - must POST", "")
		return

	}

	if parts[2] == "" {

		respond(w, 500, "Invalid Request", "")
		return

	}

	cmd = parts[2]

	file := cmdpath + "commands/cmd_" + cmd + ".txt"
	
	fp := filepath.FromSlash(file)	

	var err = os.Remove(fp)

	if err != nil {
		respond(w, 500, "Command Not Removed "+err.Error(), "")
		return
	}

	respond(w, 200, "Command Removed", "")

}

func statusHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-type", "text/plain")

	var files []string

	root := cmdpath + "commands"
	
	fp := filepath.FromSlash(root)	

	err := filepath.Walk(fp, func(path string, info os.FileInfo, err error) error {

		if path == fp {
			return nil
		}

		path = strings.Replace(path, filepath.FromSlash(root+"/cmd_"), "", -1)
		path = strings.Replace(path, filepath.FromSlash("commands/cmd_"), "", -1)
	

		parts := strings.Split(path, ".")

		if parts[0] == "commands/" {
			return nil
		}

		files = append(files, parts[0])
		return nil
	})

	if err != nil {

	}

	var payload = make(map[string]interface{})

	payload["commands"] = files

	ct := broadlink.Count()

	payload["devices_found"] = ct

	ids := broadlink.DeviceIds()

	payload["devices"] = ids

	respond(w, 200, strconv.Itoa(ct)+" Devices found", payload)

}

type Devices struct{}

func respond(w http.ResponseWriter, code int, message string, payload interface{}) {

	resp := JsonResp{
		Code:    code,
		Payload: payload,
		Message: message,
	}

	var jsonData []byte
	jsonData, err := json.Marshal(resp)

	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-type", "application/json")

	fmt.Fprintln(w, string(jsonData))

}

func loadDevices(){


	dev := getDeviceSaved()

	if(len(dev)>0){

		for k,v := range dev {



			devicetype := v[1]

			ip := v[0]

			mac := k

			deviceType, _ := strconv.Atoi(devicetype)

			broadlink.AddManualDevice(ip, mac, deviceType)

		}

	}


}

type JsonResp struct {
	Code    int         `json:"code"`
	Payload interface{} `json:"payload"`
	Message string      `json:"message"`
}

func main() {

	broadlinkgo.Logger.SetFlags(0)//disable logging
	
	var cpath=""
	
	if ( runtime.GOOS == "windows") {
		
		cpath = os.Getenv("APPDATA")+"\\broadlinkgo\\"
	
		
	}else if(runtime.GOOS == "darwin"){
		
		cpath = os.Getenv("HOME")+"/broadlinkgo/"

    }else{
	    		
		cpath = "/etc/broadlinkgo/"
	}
	
	if(cpath!=""){}
	
    flag.IntVar(&port, "port", 8000, "HTTP listener port")
	flag.StringVar(&cmdpath, "cmdpath", cpath, "Path to commands folder")
	flag.StringVar(&mode, "mode", "auto", "Auto or Manual")
	flag.Parse()
	
	log.Println("Saving commands to "+cmdpath)
	
	//bunch of windows path fixing stuff
	
	cmdpath = strings.Replace(cmdpath,"\\","/",-1)//we do this because we use filepath everywhere and need the same file path direction
	cmdpath = strings.Replace(cmdpath,"\"","",-1)
	
	slash := cmdpath[len(cmdpath)-1:]
	
	if(slash!="/"){
		
		cmdpath = cmdpath+"/"
		
	}
	
	
	ticker := time.NewTicker(5 * time.Second)
	
	if(mode=="auto"){



	go func() {
		for range ticker.C {

			broadlink = broadlinkgo.NewBroadlink()
			err := broadlink.Discover()
			if err != nil {
				log.Fatal(err)
			}

			log.Println("Found " + strconv.Itoa(broadlink.Count()) + " devices")

			if broadlink.Count() < 1 {

				log.Println("No devices found")

			} else {

				log.Println("Devices Found, updating check interval")
				ticker.Stop()
				ticker = time.NewTicker(300 * time.Second) //look every 5

			}

		}

	}()
	
	}else{
		
		broadlink = broadlinkgo.NewBroadlink()

		dev := getDeviceSaved()
	    loadDevices()
		
        	go func() {
	        	
	        	//update the ticker for the manual. mode to look for every 5 minutes
	        	ticker = time.NewTicker(300 * time.Second) //look every 5
	        	
				for range ticker.C {
	
					if broadlink.Count() < len(dev) {
		
						for k,v := range dev {
							
							if( broadlink.DeviceExists(k) ){
								
								continue
							}
							
							fmt.Println("Trying to connect to "+k)
				
							devicetype := v[1]
							
							ip := v[0]
							
							mac := k
							
							deviceType, _ := strconv.Atoi(devicetype)
			
				            broadlink.AddManualDevice(ip, mac, deviceType)
				            
				        }						
		
					}
		
				}
		
			}()    
    }




	//create cmdpath if not exist
	
	fp := filepath.FromSlash(cmdpath + "commands")

	if _, err := os.Stat(fp); os.IsNotExist(err) {
		err = os.MkdirAll(fp, 0755)
		if err != nil {
			panic(err)
		}
	}

	log.Print("Listening on port ", port)

	box := rice.MustFindBox("httpassets")
	assetsFileServer := http.StripPrefix("/assets/", http.FileServer(box.HTTPBox()))
	http.Handle("/assets/", assetsFileServer)
	http.HandleFunc("/remove/", removeHandler)
	http.HandleFunc("/removedevice/", removeDeviceHandler)
	http.HandleFunc("/device/", deviceHandler)
	http.HandleFunc("/manualdevice/", manualDeviceHandler)
	http.HandleFunc("/status/", statusHandler)
	http.HandleFunc("/cmd/", cmdHandler)
	http.HandleFunc("/discover/",func(w http.ResponseWriter, r *http.Request) {



		path := r.URL.Path
		path = strings.Replace(path, " ", "_", -1)
		path = strings.ToLower(path)

		parts := strings.Split(path,"/")

		if(len(parts)<2){

			fmt.Fprintf(w,`{"ok": "false"}`)
			return
		}

		log.Println(parts[2])

		hst := parts[2]

		err := broadlink.DiscoverHost(hst)

		if(err==nil){

			fmt.Fprintf(w,`{"ok": "true"}`)

		}else{

			fmt.Fprintf(w,`{"ok": "false"}`)

		}

	})
	http.HandleFunc("/macro/", macroHandler)
	http.HandleFunc("/learnchild/", learnChildHandler)
	http.HandleFunc("/learn/", learnHandler)
	http.HandleFunc("/", defaultHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
