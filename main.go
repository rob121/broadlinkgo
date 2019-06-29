package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
    "os"
	"github.com/kwkoo/broadlinkrm"
	"io/ioutil"
	"strconv"
	"time"
	 "path/filepath"
)

var broadlink broadlinkrm.Broadlink
var code string
var port int
var cmdpath string

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	
	fmt.Fprintln(w,"/status for command info, /learn to learn a command, /macro to send a group of commands (slash separated), cmd to send one command") 

}

func macroHandler(w http.ResponseWriter, r *http.Request) {
	
	path := r.URL.Path
	
	
	path = strings.Replace(path," ","_",-1)
	path = strings.ToLower(path)

	w.Header().Set("Content-type", "text/plain")
	
	parts := strings.Split(strings.Replace(path,"/macro/","",-1),"/")
	
	log.Println(parts)
	
	for _,v := range parts {
		
		fmt.Fprintln(w, v) 
		
		if(strings.Contains(v, ":")){
			
			cmdset := strings.Split(v,":")
			rep,_ := strconv.Atoi(cmdset[1])
			
			executeCmd(cmdset[0],rep)
		    
			
		}else{
			
			executeCmd(v,1)
		}
		
	}
	
	
	
	
}	

func executeCmd(cmd string,repeat int){
	
	
	//magic command to help macros
	if(cmd=="delay"){
		
		
		time.Sleep(1 * time.Second)
		return
		
	}
	
	
	if(repeat==0){
		repeat = 1
	}
	    
	content, err := ioutil.ReadFile(cmdpath+"commands/cmd_"+cmd+".txt")
	
	if err != nil {
		log.Println(err)
	}
	
	code = string(content)
	
	
    for i := 0; i < repeat; i++ {
     	
	broadlink.Execute("", code) 
	time.Sleep(5 * time.Millisecond) //introduce a delay here 
	
	}
	
	return 

	
}

func cmdHandler(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	
	path = strings.Replace(path," ","_",-1)
	path = strings.ToLower(path)

	
	w.Header().Set("Content-type", "text/plain")
	
	parts := strings.Split(path,"/")
	
	cmd := ""
	
    if(parts[2]!=""){
	    
	fn := parts[2]

    if(strings.Contains(fn, ":")){
			
			cmdset := strings.Split(fn,":")
			rep,_ := strconv.Atoi(cmdset[1])
			cmd = cmdset[0]
			executeCmd(cmd,rep)
		    
			
		}else{
			
			cmd = fn
			
			executeCmd(cmd,1)
	}


	
	fmt.Fprintln(w, "Command "+cmd+" executed") 
	return    
	    
    }
    
	fmt.Fprintln(w, "Command not found")	
	return	
		
	


}

func learnHandler(w http.ResponseWriter, r *http.Request) { 
	
	path := r.URL.Path
	w.Header().Set("Content-type", "text/html")
	
	path = strings.Replace(path," ","_",-1)
	path = strings.ToLower(path)
	
	r.ParseForm()
	
	
    fmt.Fprintln(w,"<h3>Learn a Device</h3>")
    fmt.Fprintln(w,"<form method='post'><label>Device Name(no spaces,lower case)</label><input name='device' value='' type='text' /><input name='submit' value='Start Learning' type='submit' /></form>")
    

	if(r.FormValue("device")!=""){
		
		dev := r.FormValue("device")
		
		fmt.Fprintln(w,"<h3>Learning device ["+dev+"]</h3>")
		
		fmt.Fprintln(w,"<h3>Press a remote button to start learning</h3>")
		
		fmt.Fprintln(w,"<iframe src='/learnchild/"+dev+"' width='100%' height='500px' ></iframe>")
		
		
	}
	
	
	
	
 
}

func learnChildHandler(w http.ResponseWriter, r *http.Request) {
	
	
	path := r.URL.Path
	
	
	path = strings.Replace(path," ","_",-1)
	path = strings.ToLower(path)
	
	w.Header().Set("Content-type", "text/plain")
	
		
		parts := strings.Split(path,"/")

		
	    if(len(parts)<3){
		
		  fmt.Fprintln(w,"Provide a button/device name")
		  return		
		  		
		}
		
		fn := ""
	 
		
		data, err := broadlink.Learn("")
		
	
		
		if err != nil {
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
	
	    if len(data) == 0 {
		  fmt.Fprintln(w, "Error: have not learned code")
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
	    
	    return
		

		

}


func statusHandler(w http.ResponseWriter, r *http.Request) {
	
	
	
	
	w.Header().Set("Content-type", "text/plain")
	
	var files []string

    root := cmdpath+"commands"
    
    err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
	    
	    if(path==root){
		    return nil
	    }
	    
	    path = strings.Replace(path,"commands/cmd_","",-1)
	    
	    parts := strings.Split(path,".")
	    
        files = append(files, parts[0])
        return nil
    })
    
    if err != nil {
       
    }
    
    for _, file := range files {
	    
	    
        fmt.Fprintln(w,file)
    }
	
	
}	


func main() {
	broadlink = broadlinkrm.NewBroadlink()
	err := broadlink.Discover()
	if err != nil {
		log.Fatal(err)
	}
	
	log.Println("Found "+strconv.Itoa(broadlink.Count())+" devices")
	
	if(broadlink.Count()<1){
	
	log.Println("No devices found")	
	 	
	}
	
	
	//port := 8081
	//cmdpath := "/etc/broadlinkrm/"

	flag.IntVar(&port, "port", 8000, "HTTP listener port")
    flag.StringVar(&cmdpath, "cmdpath","/etc/broadlinkrm/", "Path to commands folder")
	flag.Parse()

	log.Print("Listening on port ", port)
    http.HandleFunc("/status/", statusHandler)
    http.HandleFunc("/cmd/", cmdHandler)
    http.HandleFunc("/macro/", macroHandler)
    http.HandleFunc("/learnchild/", learnChildHandler)
    http.HandleFunc("/learn/", learnHandler)
	http.HandleFunc("/", defaultHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
