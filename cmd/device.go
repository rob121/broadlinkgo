package main

import(
  "time"
  "io/ioutil"
  "log"
  "errors"
  "strconv"
  "path/filepath"
  "fmt"
)

var Devices []Device

type Device struct{
  Id string `storm:"id"` //mac address
  Label string
  Ip string
  Type int
  Active bool //is the device detected, can it be communicated with?
}



func AddDevice(mac string,label string,ip string,dtype string) error{

  deviceType, _ := strconv.Atoi(dtype)
  //check it already exists in db
  d,derr := getDevice(mac)

  d.Active = false //reset it and let the broadklink system tell us it's still active
  if(derr!=nil) {

    d = Device{Id: mac, Ip: ip, Type: deviceType, Active: false}

  }

  if(len(label)>0 && label!="") {


    d.Label = label
  }

  e := db.Save(&d)

  if(e!=nil){

    log.Println(e)
  }

  if( broadlink.DeviceExists(mac) ){

    log.Println("Device already exists with broadlink")

    d.Active = true
    e = db.Save(&d)

    if(e!=nil){

      log.Println(e)
    }

    getDevices()
    return errors.New("Device Exists")

  }

  err := broadlink.AddManualDevice(ip, mac, deviceType)

  if(err!=nil){
     getDevices()
     return err
  }

  d.Active = true
  db.Save(&d)
  getDevices()

  return nil

}

func deviceTypes() map[int]string{

 return broadlink.DeviceTypes()

}

func getDevices(init ...bool) ([]Device){

    Devices = make([]Device,0)

    db.All(&Devices)


    if(len(init)>0) {
      for k := range Devices {

        Devices[k].Active = false
        Devices[k].Save()

      }
    }

    return Devices

}

func getDevice(id string) (Device,error){

  var d Device

  err := db.One("Id",id, &d)

  if(err!=nil){

     return d,err

  }

  return d,nil

}

func (d Device) Execute(cmd string,repeat int) bool{

  //magic command to help macros
  if cmd == "delay" {

    time.Sleep(1 * time.Second)
    return true

  }

  if repeat == 0 {
    repeat = 1
  }

  c,cerr := getCommandByPath(cmd)

  fp := ""

  if(cerr==nil){

  fp = c.ToFullFile()

  }else{
    //try the old fashion way, is it on disk?
    //this supports legacy install
    fp = filepath.Join(configPath,"commands",fmt.Sprintf("cmd_%s.txt",cmd))

  }

  fmt.Println(fp)


  content, err := ioutil.ReadFile(fp)

  if err != nil {
    log.Println(err)
    return false
  }

  code = string(content)

  for i := 0; i < repeat; i++ {

    broadlink.Execute(d.Id, code)
    time.Sleep(5 * time.Millisecond) //introduce a delay here

  }

  return true

}


func (d Device) Remove(){

  broadlink.RemoveDevice(d.Id)

  d.Active = false

  d.Save()

}

func (d Device) Save(){

  db.Save(&d)

}

