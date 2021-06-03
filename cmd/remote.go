package main

import(
	"log"
    "github.com/teris-io/shortid"
)

var Remotes []Remote

type Remote struct {
	Id string `storm:"id" json:"Id"`
	Label string `json:"Label"`
	Device string `json:"Device"`
	Buttons []Button `json:"Buttons"`
}

type Button struct {
	X string `json:"X"`
	Y string `json:"Y"`
	H string `json:"H"`
	W string `json:"W"`
	Label string `json:"Label"`
	Icon string `json:"Icon"`
	Color string `json:"Color"`
	Command string `json:"Command"`
}

func getRemotes() ([]Remote){




	 Remotes = make([]Remote,0)

	 err:= db.All(&Remotes)

	 if(err!=nil){

	 	log.Println(err)

	 }

	 return Remotes

}

func getRemote(id string) (Remote,error){
    var e Remote

	err:=db.One("Id",id,&e)

	if(err!=nil){
		return e,err
	}

	return e,nil

}

func addRemote(label string) (Remote){

 e:=Remote{Label: label}

 return e

}

func (e *Remote) GenerateId() (string){

	if(len(e.Id)<1) {

		e.Id, _ = shortid.Generate()

	}

	return e.Id

}

func (e *Remote) Remove() (error){


	return db.DeleteStruct(e)
}

func (e *Remote) Save() (error){
   //if it's new
    if(len(e.Id)<1){

	 e.GenerateId()

	}

	return db.Save(e)

}