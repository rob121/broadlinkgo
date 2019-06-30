package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
    "os"
	"encoding/json"
	"github.com/rob121/broadlinkgo"
	"github.com/GeertJohan/go.rice"
	"io/ioutil"
	"strconv"
	"time"
	"path/filepath"
	"text/template"
)

var broadlink broadlinkgo.Broadlink
var code string
var port int
var cmdpath string

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
	tmplMessage.Execute(w, map[string]string{"DevicesCT": strconv.Itoa(ct)})

}

func macroHandler(w http.ResponseWriter, r *http.Request) {
	
	path := r.URL.Path
	
	device := r.Form.Get("device")
	
	path = strings.Replace(path," ","_",-1)
	path = strings.ToLower(path)


	
	parts := strings.Split(strings.Replace(path,"/macro/","",-1),"/")
	
	log.Println(parts)
	
	
	status := false
	
	var state = make(map[string]bool)
	
	var output = make(map[string]interface{})
	
	for _,v := range parts {
		
		
		
		if(strings.Contains(v, ":")){
			
			cmdset := strings.Split(v,":")
			rep,_ := strconv.Atoi(cmdset[1])
			
			status = executeCmd(cmdset[0],rep,device)
			
			state[cmdset[0]]=status
		    
			
		}else{
			
			status = executeCmd(v,1,device)
			
			state[v]=status
		}
		
	}
	
     output["commands"]=state
		
     respond(w,200,"Macro executed",output)

	
	
	
	
}	

func executeCmd(cmd string,repeat int,device string) bool{
	
	
	//magic command to help macros
	if(cmd=="delay"){
		
		
		time.Sleep(1 * time.Second)
		return true
		
	}
	
	
	if(repeat==0){
		repeat = 1
	}
	    
	content, err := ioutil.ReadFile(cmdpath+"commands/cmd_"+cmd+".txt")
	
	if err != nil {
		log.Println(err)
		return false
	}
	
	code = string(content)
	
	
    for i := 0; i < repeat; i++ {
     	
	broadlink.Execute(device, code) 
	time.Sleep(5 * time.Millisecond) //introduce a delay here 
	
	}
	
	return  true

	
}

func cmdHandler(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	
	path = strings.Replace(path," ","_",-1)
	
	path = strings.ToLower(path)
	
	device := r.Form.Get("device")
	
	parts := strings.Split(path,"/")
	
	cmd := ""
	
	status := false
	
    if(parts[2]!=""){
	    
	fn := parts[2]

    if(strings.Contains(fn, ":")){
			
			cmdset := strings.Split(fn,":")
			rep,_ := strconv.Atoi(cmdset[1])
			cmd = cmdset[0]
			status = executeCmd(cmd,rep,device)
		    
			
		}else{
			
			cmd = fn
			
			status = executeCmd(cmd,1,device)
	}

    if(status==true){
	
	
	respond(w,200,"Command "+cmd+" executed","")
	return
	}
	
 
	    
    }
    respond(w,500,"Command NOT Executed","")

	return	
		
	


}

func learnHandler(w http.ResponseWriter, r *http.Request) { 
	
	path := r.URL.Path
	w.Header().Set("Content-type", "text/html")
	
	path = strings.Replace(path," ","_",-1)
	path = strings.ToLower(path)
	
	r.ParseForm()
	

    frame := ""

	if(r.FormValue("device")!=""){
		
		dev := r.FormValue("device")
		
		frame =`<h4>Learning device [`+dev+`]</h4>

		<iframe src='/learnchild/`+dev+`' width='100%' height='500px' ></iframe>`
		
		
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
	tmplMessage.Execute(w, map[string]string{"Frame": frame})
	
	
	
	
 
}

func learnChildHandler(w http.ResponseWriter, r *http.Request) {
	
	
	path := r.URL.Path
	
	
	path = strings.Replace(path," ","_",-1)
	path = strings.ToLower(path)
	
	    w.Header().Set("Content-type", "text/html")
	    fmt.Fprintln(w,"<style> body{margin:0px;padding:30px;background-color:#000;color:#FFF;}</style>")
		fmt.Fprintln(w,"<pre>")
		
	
		

		
		
		parts := strings.Split(path,"/")

		
	    if(len(parts)<3){
		
		  fmt.Fprintln(w,"Provide a button/device name")
		  fmt.Fprintln(w,"</pre>")
		  return		
		  		
		}
		
		fmt.Fprintln(w,"Waiting for remote presses<blink>....</blink>")
	    fmt.Fprintln(w,"&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<br>")
		
		fn := ""
	    
	    if f, ok := w.(http.Flusher); ok {
	     f.Flush()
		} else {
		     log.Println("no flush");
		}
		
		data, err := broadlink.Learn("")
		
	
		
		if err != nil {
			fmt.Fprintf(w, "Error: %v", err)
			fmt.Fprintln(w,"</pre>")
			return
		}
	
	    if len(data) == 0 {
		  fmt.Fprintln(w, "Error: have not learned code")
		  fmt.Fprintln(w,"</pre>")
		  return
	    }
		
		//create a file with command
		
		if(data!=""){
			
	
			
			fn = parts[2]
			
			f, err := os.Create(cmdpath+"commands/cmd_"+fn+".txt")
			 
			if err != nil {
		        fmt.Println(err)
		        return
		    }
		    
		    l , err := f.WriteString(data)
			if err != nil {
		        fmt.Println(err)
		        f.Close()
		        return
		    }
		    
		    if(l<1){}
			
			
		}
		
		fmt.Fprintln(w,"Code Detected!")
	    fmt.Fprintln(w, data)
	    fmt.Fprintln(w,"Code Saved!")
	    fmt.Fprintln(w,"Use /cmd/"+fn+" to trigger the command")  
	    fmt.Fprintln(w,"</pre>")
	    
	    return
		

		

}

func deviceHandler(w http.ResponseWriter, r *http.Request) { 
	
		
	path := r.URL.Path
	path = strings.Replace(path," ","_",-1)
	path = strings.ToLower(path)
	
	w.Header().Set("Content-type", "text/plain")
	
    parts := strings.Split(path,"/")
    
    if(len(parts)<2){
	    
	    //fmt.Fprintln(w,"Must provide a device")
	    
    }
    
    device := parts[2]
    
    fmt.Fprintln(w,device)
    
    ids := broadlink.DeviceIds()
	
    if _, ok := ids[device]; ok {
    //do something here
    //  fmt.Fprintln(w,"Device Exists")
      //update path set a from value 
      
     // fmt.Fprintln(w,r.URL.Path)
      
      r.URL.Path = strings.Replace(r.URL.Path,"/device/"+device,"",-1)
      
     // fmt.Fprintln(w,r.URL.Path)
      
      r.ParseForm()
      
      r.Form.Set("device",device)
      
      if(strings.Contains(path, "/cmd/")){
      
      cmdHandler(w,r)
      
      }else if(strings.Contains(path, "/macro/")){
	  
	  macroHandler(w,r)   
	      
      }
    
    }else{
	    
	   // fmt.Fprintln(w,"Device Does Not Exist")
	    
    }
	
}

func removeHandler(w http.ResponseWriter, r *http.Request) { 
	
	path := r.URL.Path
	
	path = strings.Replace(path," ","_",-1)
	
	path = strings.ToLower(path)
	
	parts := strings.Split(path,"/")
	
	cmd := ""
	
	if(r.Method!="POST"){
		 respond(w,500,"Invalid Request - must POST","")   
		 return
		
	}
	
	if(parts[2]==""){ 
		 
		 respond(w,500,"Invalid Request","")   
		 return
		    
	}
	
	cmd = parts[2];

    file := cmdpath+"commands/cmd_"+cmd+".txt"
	
	var err = os.Remove(file)
	
	if err != nil {		
		respond(w,500,"Command Not Removed "+err.Error(),"")
        return
 	}
	  
	
	
	respond(w,200,"Command Removed","")
	

	
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	
	
	
	
	w.Header().Set("Content-type", "text/plain")
	
	var files []string

    root := cmdpath+"commands"
    
    err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
	    
	    if(path==root){
		    return nil
	    }
	    
	
	    
	    path = strings.Replace(path,root+"/cmd_","",-1)
	    path = strings.Replace(path,"commands/cmd_","",-1)
	    
	 
	    
	    parts := strings.Split(path,".")
	    
	    if(parts[0]=="commands/"){
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
	

    respond(w,200,strconv.Itoa(ct)+" Devices found",payload)
	
	
}	

type Devices struct{}


func respond(w http.ResponseWriter,code int,message string,payload interface{}){
	
	resp := JsonResp{
		Code: code,
		Payload: payload,
		Message: message,		
	}
	
	var jsonData []byte
    jsonData, err := json.Marshal(resp)
    
    if err != nil {
    log.Println(err)
    }
    
    w.Header().Set("Content-type", "application/json")

    fmt.Fprintln(w,string(jsonData))
	
	
}

type JsonResp struct{
 Code int `json:"code"`
 Payload interface{} `json:"payload"`
 Message string `json:"message"`
}



func main() {
	
	
	ticker := time.NewTicker(5 * time.Second)
	
	go func() {
        for  range ticker.C {
	
			broadlink = broadlinkgo.NewBroadlink()
			err := broadlink.Discover()
			if err != nil {
				log.Fatal(err)
			}
			
			log.Println("Found "+strconv.Itoa(broadlink.Count())+" devices")
			
			if(broadlink.Count()<1){
			
			log.Println("No devices found")	

	 	
			}else{
				
			log.Println("Devices Found, updating check interval")
			ticker.Stop()
			ticker = time.NewTicker(300 * time.Second)//look every 5
				
			}
			
	 	}		
	
	}()
	
    
	
	
	//port := 8081
	//cmdpath := "/etc/broadlinkgo/"

	flag.IntVar(&port, "port", 8000, "HTTP listener port")
    flag.StringVar(&cmdpath, "cmdpath","/etc/broadlinkgo/", "Path to commands folder")
	flag.Parse()
	
	
	//create cmdpath if not exist
	
	if _, err := os.Stat(cmdpath+"commands"); os.IsNotExist(err) {
              err = os.MkdirAll(cmdpath+"commands", 0755)
              if err != nil {
                      panic(err)
              }
      }

	log.Print("Listening on port ", port)
	
	box := rice.MustFindBox("httpassets")
    assetsFileServer := http.StripPrefix("/assets/", http.FileServer(box.HTTPBox()))
    http.Handle("/assets/", assetsFileServer)
    http.HandleFunc("/remove/", removeHandler)
	http.HandleFunc("/device/", deviceHandler)
    http.HandleFunc("/status/", statusHandler)
    http.HandleFunc("/cmd/", cmdHandler)
    http.HandleFunc("/macro/", macroHandler)
    http.HandleFunc("/learnchild/", learnChildHandler)
    http.HandleFunc("/learn/", learnHandler)
	http.HandleFunc("/", defaultHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
