package main

import(
	"encoding/json"
	"net/http"
	"log"
	"fmt"
	"embed"
	"html/template"
	"github.com/gorilla/mux"
	"time"
	"strconv"
	"io/ioutil"
	"path/filepath"
	"os"
)
//go:embed assets/*
var assetsfs embed.FS
type PageData map[string]interface{}

type JsonResp struct {
	Code    int         `json:"code"`
	Payload interface{} `json:"payload"`
	Message string      `json:"message"`
}

func startHTTPServer(){


	r:= mux.NewRouter()

	if(fileExists(assets_dir)){
		log.Println("Assets found on disk")
		//TODO fix basepath here
		r.PathPrefix(fmt.Sprintf("/%s/",assets_dir)).Handler(http.StripPrefix(fmt.Sprintf("/%s/",assets_dir), http.FileServer(http.Dir(assets_dir))))
	}else{
		log.Println("Using embedded assets")
		r.PathPrefix("/assets/").Handler(http.FileServer(http.FS(assetsfs)))
	}

	//

	r.HandleFunc("/learn/{id}", httpLearnHandler)
	r.HandleFunc("/learn", httpLearnHandler)
	r.HandleFunc("/learncode", httpLearnCodeHandler)

	r.HandleFunc("/commands{format:[\\.json]*}", httpCommandsHandler)
	r.HandleFunc("/command/{id}", httpCommandHandler)
	r.HandleFunc("/commandremove/{id}", httpCommandRemoveHandler)
	r.HandleFunc("/commandsave", httpCommandSaveHandler)
	r.HandleFunc("/macros{format:[\\.json]*}", httpMacrosHandler)
	r.HandleFunc("/macro/{id}", httpMacroHandler)
	r.HandleFunc("/macroremove/{id}", httpMacroRemoveHandler)
	r.HandleFunc("/macrosave", httpMacroSaveHandler)
	r.HandleFunc("/deviceedit/{id}", httpDeviceHandler)
	r.HandleFunc("/devicesave", httpSaveDeviceHandler)
	r.HandleFunc("/equipments{format:[\\.json]*}", httpEquipmentsHandler)
	r.HandleFunc("/equipment/{id}", httpEquipmentHandler)
	r.HandleFunc("/equipmentremove/{id}", httpEquipmentRemoveHandler)
	r.HandleFunc("/equipmentsave", httpEquipmentSaveHandler)
	r.HandleFunc("/remotes{format:[\\.json]*}", httpRemotesHandler)
	r.HandleFunc("/remotesavebuttons",httpRemoteSaveButtonsHandler)
	r.HandleFunc("/remote/{id}", httpRemoteHandler)
	r.HandleFunc("/universal/{id}", httpUniversalHandler)
	r.HandleFunc("/remoteremove/{id}", httpRemoteRemoveHandler)
	r.HandleFunc("/remotesave", httpRemoteSaveHandler)
	r.HandleFunc("/ping", httpHealthHandler)
	r.HandleFunc("/api/status", apiStatusHandler)
	r.HandleFunc("/api/remotes", apiRemotesHandler)
	r.HandleFunc("/cmd/{cmds}", apiCmdHandler)
	r.HandleFunc("/device/{device}/{type}/{values:[a-zA-Z0-9@:=\\+_\\.\\-\\/]+}", apiDeviceHandler)
	r.HandleFunc("/macro/{cmds}", apiMacroHandler)
	r.HandleFunc("/index{format:[\\.json]*}", httpDefaultHandler)
	r.HandleFunc("/", httpDefaultHandler)
	http.Handle("/", r)

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf(":%s", port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Listening on port", port)
	log.Fatal(srv.ListenAndServe())


}

func loadTmpl(dest string) (*template.Template){

	tmpl := template.New("default")
	tmpl, _ = tmpl.New("header").ParseFS(assetsfs,"assets/tmpl/header.html")
	tmpl, _ = tmpl.New("footer").ParseFS(assetsfs,"assets/tmpl/footer.html")

	var err error

	tmpl, err = tmpl.New(filepath.Base(dest)).Funcs(template.FuncMap{
		"getTimestamp": func() int64 {
			return time.Now().Unix()
		},
	}).ParseFS(assetsfs,dest)

	if(err!=nil){

		log.Println(err)
	}



	//log.Printf("%#v\n",tmpl.DefinedTemplates())

    return tmpl
}

func httpMacrosHandler(w http.ResponseWriter, r *http.Request) {

	tmpl := loadTmpl("assets/tmpl/macros.html")

	data := PageData{}

	data["Title"] = "Macros"

	if(r.Header.Get("X-Requested-With") == "xmlhttprequest" || r.Header.Get("Hx-Request")=="true") {

		data["Macros"] = getMacros()

		respond(w,200,"OK",data)
		return
	}

	tmpl.Execute(w, data)

}

func httpMacroHandler(w http.ResponseWriter, r *http.Request) {

	vars:=mux.Vars(r)

	id:=vars["id"]

	var re Macro
	var err error


	if(id!="new" && id!="0" && id!=""){
		re,err = getMacro(id)

		if(err!=nil){
			//storm.ErrNotFound
			re = Macro{}
			log.Println(err)

		}
	}

	tmpl := loadTmpl("assets/tmpl/macro.html")

	data := PageData{}

	coms := getCommands()

	data["Title"] = "Macro"
	data["Macro"] = re
	data["IconSelect"] = template.HTML(getIconSelect())
	data["Commands"] = coms

	device_sel := getDeviceSelect(re.Device,"")

	data["DeviceSelect"] = template.HTML(device_sel)

	e:=tmpl.Execute(w, data)


	if(e!=nil){

		log.Println(e)
	}

}




func httpMacroSaveHandler(w http.ResponseWriter, r *http.Request) {

	var rb Macro
	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&rb)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("%#v",rb)

	if(rb.Id==""){
		rb.Id = rb.GenerateId()
	}

	log.Printf("%#v",rb)

	e:=db.Save(&rb)

	if(e!=nil){
		log.Println(e)
	}



}
func httpMacroRemoveHandler(w http.ResponseWriter, r *http.Request) {


	w.Header().Set("Content-type", "application/json")

	r.ParseForm()

	if r.Method != "POST" {
		respond(w, 500, "Invalid Request - must POST", "")
		return
	}

	vars := mux.Vars(r)

	var id string
	var ok bool

	if id,ok = vars["id"]; ok {

		e,err := getMacro(id)

		if(err==nil){

			rerr := e.Remove()

			if(rerr!=nil){

				log.Println(rerr)
			}
			w.Header().Set("HX-Refresh", "true")
			respond(w, 200, "Success, removed", "")
			return

		}else{

			log.Println("Item not found")
		}

	}

	respond(w, 500, "General Error", "")
	return

}

func httpRemotesHandler(w http.ResponseWriter, r *http.Request) {

	vars:=mux.Vars(r)

	format := vars["format"]

	tmpl := loadTmpl("assets/tmpl/remotes.html")

	data := PageData{}

	data["Title"] = "Remotes"

	if(format==".json" || (r.Header.Get("X-Requested-With") == "xmlhttprequest" || r.Header.Get("Hx-Request")=="true")) {

		data["Remotes"] = getRemotes()
	
		respond(noCache(w),200,"OK",data)
		return
	}

	tmpl.Execute(w, data)

}

func httpRemoteSaveButtonsHandler(w http.ResponseWriter, r *http.Request) {

	var rb Remote

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&rb)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	e:=db.Save(&rb)

	if(e!=nil){
		log.Println(e)
	}

}

func httpUniversalHandler(w http.ResponseWriter, r *http.Request) {

	vars:=mux.Vars(r)

	id:=vars["id"]

	var re Remote
	var err error


	if(id!="new" && id!="0"){
		re,err = getRemote(id)

		if(err!=nil){
			//storm.ErrNotFound
			re = Remote{}
			log.Println(err)

		}
	}

	tmpl := loadTmpl("assets/tmpl/universal.html")

	data := PageData{}

	data["Title"] = "Remotes"
	data["Remote"] = re
	data["Commands"] = getCommands()

	for k,b := range re.Buttons {

		log.Println(b.Command)

		c,e:=getCommand(b.Command)

		if(e==nil){

			re.Buttons[k].Command = fmt.Sprintf("/device/%s/cmd/%s",re.Device,c.ToCmd())

		}else{

			m,e:= getMacro(b.Command)

			if(e==nil) {

				re.Buttons[k].Command = fmt.Sprintf("/device/%s/macro/%s", re.Device, m.ToCmd())

			}
		}

	}

	device_sel := getDeviceSelect(re.Device,"")

	data["DeviceSelect"] = template.HTML(device_sel)

	e:=tmpl.Execute(w, data)


	if(e!=nil){

		log.Println(e)
	}

}

func httpRemoteHandler(w http.ResponseWriter, r *http.Request) {

	vars:=mux.Vars(r)

	id:=vars["id"]

	var re Remote
    var err error


	if(id!="new" && id!="0"){
		re,err = getRemote(id)

		if(err!=nil){
			//storm.ErrNotFound
			re = Remote{}
			log.Println(err)

		}
	}

	tmpl := loadTmpl("assets/tmpl/remote.html")

	data := PageData{}

	data["Title"] = "Remotes"
	data["Remote"] = re
	data["Commands"] = getCommands()
	data["Macros"] = getMacros()

	device_sel := getDeviceSelect(re.Device,"")

	data["DeviceSelect"] = template.HTML(device_sel)

	e:=tmpl.Execute(w, data)


    if(e!=nil){

    	log.Println(e)
	}

}
func httpRemoteRemoveHandler(w http.ResponseWriter, r *http.Request) {



	w.Header().Set("Content-type", "application/json")

	r.ParseForm()

	if r.Method != "POST" {
		respond(w, 500, "Invalid Request - must POST", "")
		return
	}

	vars := mux.Vars(r)

	var id string
	var ok bool

	if id,ok = vars["id"]; ok {

		e,err := getRemote(id)

		if(err==nil){

			rerr := e.Remove()

			if(rerr!=nil){

				log.Println(rerr)
			}
			w.Header().Set("HX-Refresh", "true")
			respond(w, 200, "Success, removed", "")
			return

		}else{

			log.Println("Item not found")
		}

	}

	respond(w, 500, "General Error", "")
	return

}

func httpRemoteSaveHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		respond(w, 500, "Invalid Request - must POST", "")
		return
	}

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	id := r.Form.Get("id")
	label := r.Form.Get("label")
	device := r.Form.Get("device")

	var re Remote
	var err error

	if(id!="new" && id!="") {


		re,err = getRemote(id)

	}

	re.Label = label
	re.Device = device

	if(err==nil){

		e:=re.Save()
		if(e!=nil){
			log.Println(e)
			respond(w,500,"Error",err.Error())
			return
		}
	}else{
		log.Println(err)

		respond(w,500,"Error",err.Error())
		return
	}

	log.Printf("%#v",re)

	 w.Header().Set("HX-Redirect",fmt.Sprintf("/remote/%s",re.Id))
	respond(w,200,"OK","")
	return


}


func httpHealthHandler(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "pong")
}

func httpLearnCodeHandler(w http.ResponseWriter, r *http.Request) {


	if r.Method != "POST" {
		respond(w, 500, "Invalid Request - must POST", "")
		return
	}

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	w.Header().Set("Content-type", "text/html")
	fmt.Fprintln(w, "<style> body{font-family: Consolas,monospace;margin:0px;padding:30px;background-color:#000;color:#FFF;}</style>")




	device := r.Form.Get("device")
	label := r.Form.Get("label")
	method := r.Form.Get("method")
	color := r.Form.Get("color")
	icon := r.Form.Get("icon")
	equipment := r.Form.Get("equipment")

	if(label==""){

		fmt.Fprintln(w,"Unable to continue: label missing")

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		} else {
			log.Println("no flush")
		}
		return

	}

	e,gerr := getEquipment(equipment)


	if(gerr!=nil){

		fmt.Fprintln(w,"Equipment Error:"+gerr.Error())

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		} else {
			log.Println("no flush")
		}
		return

	}

	cmd := addCommand(label,icon,e,color)

	fp:= cmd.ToFullPath()
	ff:= cmd.ToFullFile()
   //TODO some error handling here
	os.MkdirAll(fp, os.ModePerm)

	if (method=="rf") {
		fmt.Fprintln(w, `Waiting for RF remote. IMPORTANT - press on for 1 second and release until learning is finished<span class="blink">....</span>`)
	}else{
		fmt.Fprintln(w, `Waiting for ir remote presses<span class="blink">....</span>`)
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

	if (method=="rf") {

		data, err = broadlink.LearnRF(device)
	}else{
		data, err = broadlink.Learn(device)
	}

	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)

		return
	}

	if len(data) == 0 {
		fmt.Fprintln(w, "Error: have not learned code")

		return
	}else{

		err := ioutil.WriteFile(ff,[]byte(data), 0644)
		if(err!=nil){}


		cmd.Save()

	}

	fmt.Fprintln(w, "<br>Code Detected!:")
	fmt.Fprintln(w, `<br><div style="max-width:700px;word-wrap:break-word;">`)
	fmt.Fprintln(w, data)
	fmt.Fprintln(w, "</div>")
	fmt.Fprintln(w, "<br>Code Saved to: "+ff)

	if(device!=""){

		fmt.Fprintln(w, "<br>Use /device/"+device+"/cmd/"+fn+" to trigger the command")

	}else{

		fmt.Fprintln(w, "<br>Use /cmd/"+fn+" to trigger the command")

	}

	return

}

func httpLearnHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id := vars["id"]

	tmpl := loadTmpl("assets/tmpl/learn.html")

	data := PageData{}

	data["Title"] = "Learning"

	equipment_sel := "<select class='form-control' name='equipment' >"

	ke := getEquipments()


	log.Printf("Id lookup %s\n",id)

	for _, v := range ke {



		equipment_sel += "<option value='" + v.Id + "' >" + fmt.Sprintf("%s-%s (%s)",v.Manufacturer,v.Model,v.Label) + "</option>"

	}

	equipment_sel += "</select>"

	data["LearnEquipment"] = template.HTML(equipment_sel)

	device_sel := "<select class='form-control' name='device' >"

	kd := getDevices()

	s := ""

	for _, v := range kd {

		if(id==v.Id){
			s = "selected='selected'"
		}

		device_sel += "<option "+s+" value='" + v.Id + "' >" + fmt.Sprintf("%s (%s/%s)",v.Label,v.Ip,v.Id)  + "</option>"
		s = ""
	}

	device_sel += "</select>"

	icon_sel := getIconSelect()

	data["IconSelect"] = template.HTML(icon_sel)

	data["LearnDevice"] = template.HTML(device_sel)

	tmpl.Execute(w, data)


}

func httpEquipmentsHandler(w http.ResponseWriter, r *http.Request) {


	tmpl := loadTmpl("assets/tmpl/equipments.html")


	data := PageData{}

	data["Title"] = "Equipment"


	if(r.Header.Get("X-Requested-With") == "xmlhttprequest" || r.Header.Get("Hx-Request")=="true") {

		data["Equipments"] = getEquipments()

		respond(noCache(w),200,"OK",data)
		return
	}

	tmpl.Execute(w, data)

}

func httpEquipmentHandler(w http.ResponseWriter, r *http.Request) {

vars := mux.Vars(r)

id := vars["id"]

e := Equipment{}

if(id!="new"){

	e,_ = getEquipment(id)

}


tmpl := loadTmpl("assets/tmpl/equipment.html")


data := PageData{}

data["Title"] = "Equipment: "+id

data["Equipment"] = e




err := tmpl.Execute(w, data)

if(err!=nil){

log.Println(err)
}


}

func httpEquipmentRemoveHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-type", "application/json")

	r.ParseForm()

	if r.Method != "POST" {
		respond(w, 500, "Invalid Request - must POST", "")
		return
	}

	vars := mux.Vars(r)

	var id string
	var ok bool

	if id,ok = vars["id"]; ok {

		e,err := getEquipment(id)

		if(err==nil){

			rerr := e.Remove()

			if(rerr!=nil){

				log.Println(rerr)
			}
			w.Header().Set("HX-Refresh", "true")
			respond(w, 200, "Success, removed", "")
            return

		}else{

			log.Println("Equipment Item not found")
		}

	}

	respond(w, 500, "General Error", "")
	return

}

func httpEquipmentSaveHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-type", "application/json")

	r.ParseForm()

	if r.Method != "POST" {
		respond(w, 500, "Invalid Request - must POST", "")
		return

	}


	id := r.FormValue("id")

	label := r.FormValue("label")

	man := r.FormValue("manufacturer")

	model := r.FormValue("model")

	if label == "" || man == "" || model == ""{

		respond(w, 500, "Label, Manufacturer,Model Required", "")
		return
	}

	var e Equipment

	if(id=="new" || id=="") {

		e = addEquipment(label, man, model)

		e.GenerateId()


	}else{

		e,_ = getEquipment(id)

		e.Label = label
		e.Manufacturer = man
		e.Model = model

	}

    state := e.Save()


	if state != nil {

		respond(w, 500, "Add Error: "+state.Error(), "")
		return

	}


	w.Header().Set("HX-Redirect",fmt.Sprintf("/equipment/%s",e.Id))

	respond(w, 200, "Equipment Added Succesfully", "")

}

func httpCommandsHandler(w http.ResponseWriter, r *http.Request) {

	tmpl := loadTmpl("assets/tmpl/commands.html")

	data := PageData{}

	data["Title"] = "Home"

	if(r.Header.Get("X-Requested-With") == "xmlhttprequest" || r.Header.Get("Hx-Request")=="true") {
		data["Commands"] = getCommands()
		respond(noCache(w),200,"OK",data)
		return
	}


	tmpl.Execute(noCache(w), data)

}

func httpCommandSaveHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-type", "application/json")

	r.ParseForm()

	if r.Method != "POST" {
		respond(w, 500, "Invalid Request - must POST", "")
		return

	}


	id := r.FormValue("id")

	label := r.FormValue("label")

	color := r.FormValue("color")



	icon := r.FormValue("icon")

	if label == "" {

		respond(w, 500, "Label Required", "")
		return
	}

    cmd,e := getCommand(id)

    if(e==nil) {


    	//equipment can't be changed so only set if it's empty
    	if(cmd.Equipment==Equipment{}){
		  eq,err := getEquipment(r.FormValue("equipment"))
	      if(err==nil){
	      	cmd.Equipment = eq
	      	cmd.Path = cmd.ToShortPath()
		  }
		}
		cmd.Label = label
		cmd.Icon = icon
		cmd.Color = color
		state := cmd.Save()

		if state != nil {

			respond(w, 500, "Add Error: "+state.Error(), "")
			return

		}


		w.Header().Set("HX-Redirect",fmt.Sprintf("/command/%s",cmd.Id))

		respond(w, 200, "Equipment Added Succesfully", "")
		return

	}

	respond(w, 500, "Add Error", "")



}

func httpCommandHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id := vars["id"]

	c := Command{}
    var cerr error

	if(id!="new"){
		c,cerr = getCommand(id)

		if(cerr!=nil){

			log.Println(cerr)

		}

	}

	tmpl := loadTmpl("assets/tmpl/command.html")

	

	data := PageData{}

	data["Title"] = "Equipment: "+id

	//because we store a static instance of equipment, refresh it on load
	c.EquipmentRefresh()

	data["Command"] = c

	data["IconSelect"] = template.HTML(getIconSelect())

	data["DeviceSelect"] = template.HTML(getDeviceSelect("","device"))

	data["EquipmentSelect"] = template.HTML(getEquipmentSelect("","equipment"))


	err := tmpl.Execute(w, data)

	if(err!=nil){

		log.Println(err)
	}


}

func httpCommandRemoveHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-type", "application/json")

	r.ParseForm()

	if r.Method != "POST" {
		respond(w, 500, "Invalid Request - must POST", "")
		return
	}

	vars := mux.Vars(r)

	var id string
	var ok bool

	if id,ok = vars["id"]; ok {

		e,err := getCommand(id)

		if(err==nil){

			e.Remove()
			w.Header().Set("HX-Refresh", "true")
			respond(w, 200, "Success, removed", "")
			return

		}

	}

	respond(w, 500, "General Error", "")
	return

}

func httpSaveDeviceHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-type", "application/json")

	r.ParseForm()

	if r.Method != "POST" {
		respond(w, 500, "Invalid Request - must POST", "")
		return

	}

	ip := r.FormValue("ip")

	id := r.FormValue("id")

	mac := r.FormValue("id")

	label := r.FormValue("label")

	devicetype := r.FormValue("devicetype")

	if ip == "" || mac == "" {

		respond(w, 500, "Ip, Mac Required", "")
		return
	}

	if(len(id)>0){

		dev,err := getDevice(id)

		if(err==nil){

			dev.Label = label

			dt,_ := strconv.Atoi(devicetype)
			dev.Type = dt
			dev.Save()
			respond(w, 200, "Device Saved", "")
			return


		}

	}else {

		state := AddDevice(mac, label, ip, r.Form.Get("devicetype"))

		if state != nil {

			respond(w, 500, "Add Error "+state.Error(), "")
			return

		}

	}

	respond(w, 200, "Device Added Succesfully", "")

}

func httpDeviceHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id := vars["id"]

	if(len(id)<1){

		fmt.Fprintf(w,"ERROR 500")
		return

	}

	tmpl := loadTmpl("assets/tmpl/device.html")


	var err error
	var d Device


	if(id!="new"){


		d,err = getDevice(id)

		if(err!=nil){

			log.Println(err)
		}


	}




	data := PageData{}

	kd := deviceTypes()

	device_sel := "<select class='form-control' name='devicetype' >"

	kk := ""

	for k, v := range kd {

		kk = strconv.Itoa(k)

		sel:=""

		if(k==d.Type){

        sel = `selected="selected"`

		}

		device_sel += "<option "+sel+" value='" + kk + "' >" + v + "</option>"

	}

	device_sel += "</select>"


	data["Title"] = "Device: "+id
	data["Device"] = d
	data["DeviceSelect"] = template.HTML(device_sel)

	err = tmpl.Execute(w, data)

	if(err!=nil){

		log.Println(err)
	}


}

func httpDefaultHandler(w http.ResponseWriter, r *http.Request) {

	    tmpl := loadTmpl("assets/tmpl/index.html")

		data := PageData{}

		data["Title"] = "Home"

		if(r.Header.Get("X-Requested-With") == "xmlhttprequest" || r.Header.Get("Hx-Request")=="true") {

			data["Devices"] = getDevices()
			data["Types"] = deviceTypes()

			respond(noCache(w),200,"OK",data)
			return
		}


		tmpl.Execute(w, data)



}

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


func getIconSelect() (string){

	icon_sel:=`

<details class="icon_sel form-control">
<summary>Icon</summary>
<p>
  <a href="#" ><i class="fas fa-play"></i> Play</a>
  <a href="#" ><i class="fas fa-pause"></i> Pause</a>
  <a href="#" ><i class="fas fa-stop"></i> Stop</a>
  <a href="#" ><i class="fas fa-circle"></i> Record</a>
  <a href="#" ><i class="fas fa-info-circle"></i> Info</a>
  <a href="#" ><i class="fas fa-check-circle"></i> Check</a>
  <a href="#" ><i class="fas fa-power-off"></i> Power</a>
  <a href="#" ><i class="fas fa-times-circle"></i> Exit</a>
  <a href="#" ><i class="fas fa-bars"></i> Menu</a>
  <a href="#" ><i class="fas fa-undo-alt"></i> Return</a>
  <a href="#" ><i class="fas fa-forward"></i> Forward</a>
  <a href="#" ><i class="fas fa-fast-forward"></i> Fast Forward</a>
  <a href="#" ><i class="fas fa-fast-backward"></i> Backward</a>  
  <a href="#" ><i class="fas fa-fast-backward"></i> Fast Backward</a>
  <a href="#" ><i class="fas fa-volume-up"></i> Volume Up</a>
  <a href="#" ><i class="fas fa-volume-down"></i> Volume Down</a>
  <a href="#" ><i class="fas fa-volume-mute"></i> Mute</a>
  <a href="#" ><i class="fas fa-plus"></i> Plus</a>
  <a href="#" ><i class="fas fa-minus"></i> Minus</a>
  <a href="#" ><i class="fas fa-plus-square"></i> Plus</a>
  <a href="#" ><i class="fas fa-minus-square"></i> Minus</a>
  <a href="#" ><i class="fas fa-num-1"></i> 1 Button</a>
  <a href="#" ><i class="fas fa-num-2"></i> 2 Button</a>
  <a href="#" ><i class="fas fa-num-3"></i> 3 Button</a>
  <a href="#" ><i class="fas fa-num-4"></i> 4 Button</a>
  <a href="#" ><i class="fas fa-num-5"></i> 5 Button</a>
  <a href="#" ><i class="fas fa-num-6"></i> 6 Button</a>
  <a href="#" ><i class="fas fa-num-7"></i> 7 Button</a>
  <a href="#" ><i class="fas fa-num-8"></i> 8 Button</a>
  <a href="#" ><i class="fas fa-num-9"></i> 9 Button</a>
  <a href="#" ><i class="fas fa-num-0"></i> 0 Button</a>
  <a href="#" ><i class="fas fa-remote-menu"></i> Menu</a>
  <a href="#" ><i class="fas fa-remote-exit"></i> Exit</a>
  <a href="#" ><i class="fas fa-tv"></i> TV</a>
  <a href="#" ><i class="fas fa-laptop"></i> Computer</a>
  <a href="#" ><i class="fas fa-server"></i> Component</a>
  <a href="#" ><i class="fas fa-hdd"></i> Device</a>
  <a href="#" ><i class="fas fa-tools"></i> Tools</a>
  <a href="#" ><i class="fas fa-caret-up"></i> Up</a>
  <a href="#" ><i class="fas fa-caret-down"></i> Down</a>
  <a href="#" ><i class="fas fa-caret-left"></i> Left</a>
  <a href="#" ><i class="fas fa-caret-right"></i> Right</a>
  <a href="#" ><i class="fas fa-music"></i> Music</a>
  <a href="#" ><i class="fas fa-film"></i> Movie</a>
  <a href="#" ><i class="fas fa-youtube"></i> Youtube</a>

</p>
</details>
`

	return icon_sel

}


func getEquipmentSelect(selected string,id string) (string){

	if(id==""){

		id="equipment"
	}

	sel := "<select class='form-control' id='"+id+"' name='equipment' >"

	sel += "<option value='' >-- Select Equipment --</option>"

	kd := getEquipments()

	s := ""

	for _, v := range kd {

		if(selected == v.Id){

			s = "selected='selected'"

		}

		sel += "<option "+s+" value='" + v.Id + "' >" + fmt.Sprintf("%s (%s/%s)",v.Label,v.Manufacturer,v.Model)  + "</option>"
		s = ""
	}

	sel += "</select>"

	return sel


}


func getDeviceSelect(selected string,id string) (string){

	if(id==""){

		id="device"
	}

	device_sel := "<select class='form-control' id='"+id+"' name='device' >"

	device_sel += "<option value='' >-- Select Target Device --</option>"

	kd := getDevices()

	sel := ""

	for _, v := range kd {

		if(selected == v.Id){

			sel = "selected='selected'"

		}

		device_sel += "<option "+sel+" value='" + v.Id + "' >" + fmt.Sprintf("%s (%s/%s)",v.Label,v.Ip,v.Id)  + "</option>"
		sel = ""
	}

	device_sel += "</select>"

	return device_sel


}


func noCache(w http.ResponseWriter) (http.ResponseWriter){

	var epoch = time.Unix(0, 0).Format(time.RFC1123)

	var noCacheHeaders = map[string]string{
		"Expires":         epoch,
		"Vary":            "Accept",
		"Cache-Control":   "no-cache, private, max-age=0",
		"Pragma":          "no-cache",
		"X-Accel-Expires": "0",
	}

	// Set our NoCache headers
	for k, v := range noCacheHeaders {
		w.Header().Set(k, v)
	}

	return w

}