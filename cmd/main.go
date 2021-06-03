package main

import(
	"fmt"
	"flag"
	"github.com/rob121/broadlinkgo"
	"github.com/asdine/storm/v3"
    "github.com/kirsle/configdir"
	"path/filepath"
	"os"
	"log"
)

var broadlink broadlinkgo.Broadlink
var code string
var configPath string
var db *storm.DB
var port string
var cmdpath string
var mode string
var debug bool
var assets_dir string
var upgrade bool

func main(){

	var err error
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.StringVar(&assets_dir, "assets_dir","", "Use your own files instead of embedded files")
	flag.BoolVar(&debug, "debug",false, "Turn on debugging")
	flag.BoolVar(&upgrade, "upgrade",false, "Run upgrade")
	flag.StringVar(&port, "port", "8000", "HTTP listener port")
	flag.StringVar(&mode, "mode", "auto", "Auto or Manual")
	flag.Parse()

	setupConfigPath()


	//open the database for devices

	db, err = storm.Open(filepath.Join(configPath, "devices.db"))

	if(err!=nil){

		panic("Unable to open device database "+err.Error())

	}

	if(upgrade!=false) {

		bootstrapOldCommands()

	}

	defer db.Close()

   //populate the namespace with saved devices for speed
	getDevices(true)


	//enable logging in the lib

	if(debug==false){

		broadlinkgo.Logger.SetFlags(1)

	}


	broadlink = broadlinkgo.NewBroadlink() //global handler

    go broadlinkDeviceWatch()
	go startHTTPServer()

	select{}
}

func setupConfigPath(){

	configPath = configdir.LocalConfig("broadlinkgo")
	err := configdir.MakePath(configPath) // Ensure it exists.

	if err != nil {
		panic(err)
	}

	err = configdir.MakePath(filepath.Join(configPath,"commands"))// Ensure it exists.

	if err != nil {
		panic(err)
	}


	fmt.Printf("Config data saved to: %s\n",configPath)

}

func fileExists(path string) bool{

	if _, err := os.Stat(path); err == nil {
		// path/to/whatever exists

		return true

	} else if os.IsNotExist(err) {
		return false

	} else {
		return false

	}


}

