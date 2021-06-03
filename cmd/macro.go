package main

import(
	"log"
	"fmt"
	"strings"
	"sort"
	"github.com/teris-io/shortid"
)

var Macros []Macro

type Macro struct {
	Id string `storm:"id" json:"Id"`
	Label string `json:"Label"`
	Icon string
	Device string
	Color string
	Commands []MacroCommand `json:"Commands"`
}

type MacroCommand struct{
	Order int
	Repeat int
	Type string
	Command string
}


func getMacros() ([]Macro){

	Macros = make([]Macro,0)

	err:= db.All(&Macros)

	if(err!=nil){

		log.Println(err)

	}

	return Macros

}

func(mc *MacroCommand) GetCommand()(Command,error){

	return getCommand(mc.Command)

}

func getMacro(id string) (Macro,error){
	var e Macro

	err:=db.One("Id",id,&e)

	if(err!=nil){
		return e,err
	}

	return e,nil

}

func addMacro(label string) (Macro){

	e:=Macro{Label: label}

	return e

}

func (e *Macro) GenerateId() (string){

	if(len(e.Id)<1) {

		e.Id, _ = shortid.Generate()

	}

	return e.Id

}


func (e *Macro) ToCmd() (string){

	sort.Slice(e.Commands, func(i, j int) bool {
		return e.Commands[i].Order < e.Commands[j].Order
	})

	var cmd []string

	for _,v := range e.Commands {

         c := ""

         cc,err := v.GetCommand()

		if(err!=nil){

			continue
		}

		 if(v.Repeat>0){

		 	c = fmt.Sprintf("%s:%d",cc.ToCmd(),v.Repeat)

		 }else{

		 	c = fmt.Sprintf("%s",cc.ToCmd())
		 }


		 cmd = append(cmd,c)

	}

	return fmt.Sprintf("%s",strings.Join(cmd,"/"))



}

func (e *Macro) Remove() (error){


	return db.DeleteStruct(e)
}

func (e *Macro) Save() (error){
	//if it's new
	if(len(e.Id)<1){

		e.GenerateId()

	}

	return db.Save(e)

}