package main

import(
	"regexp"
	"fmt"
	"strings"
)

var Equipments []Equipment

type Equipment struct {
	Id string `storm:"id"`
	Label string
	Manufacturer string
	Model string
}

func getEquipments() ([]Equipment){

	 Equipments = make([]Equipment,0)

	 db.All(&Equipments)

	 return Equipments

}


func getEquipment(id string) (Equipment,error){
    var e Equipment

	err:=db.One("Id",id,&e)

	if(err!=nil){
		return e,err
	}

	return e,nil

}

func addEquipment(label string, man string, model string) (Equipment){

 e:=Equipment{Label: label,Manufacturer: man,Model:model}

 e.GenerateId()

 return e

}

func (e *Equipment) GenerateId(){



	if(len(e.Id)<1){

		re := regexp.MustCompile(`\W`)
		man := fmt.Sprintf("%s",re.ReplaceAll([]byte(e.Manufacturer),[]byte("")))
		mod := fmt.Sprintf("%s",re.ReplaceAll([]byte(e.Model),[]byte("")))
		e.Id = strings.ToLower(fmt.Sprintf("%s_%s",man,mod))

	}





}

func (e *Equipment) Refresh() (Equipment){


	nw,err := getEquipment(e.Id)

	if(err!=nil){}

	e = &nw

	return nw


}

func (e *Equipment) Remove() (error){


	return db.DeleteStruct(e)
}

func (e *Equipment) Save() (error){


	return db.Save(e)

}