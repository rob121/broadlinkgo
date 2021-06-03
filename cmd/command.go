package main


import(
	"os"
	"fmt"
	"path/filepath"
    "log"
	"regexp"
	"strings"
	"hash/adler32"
	"github.com/speps/go-hashids/v2"
)

type Command struct{
	Id string `storm:"id"`
	Label string
	Icon string
	Equipment Equipment
	Color string
	Path string //partial path
}

func addCommand(label string,icon string,equipment Equipment,color string) (Command){

	c := Command{Label: label,Icon: icon,Equipment: equipment,Color: color}

	c.Path = c.ToShortPath()

	return c

}

func getCommand(id string) (Command,error){

	var c Command

	if(id=="delay"){

		c = Command{Id:"delay",Label: "Delay",Icon: "fa-stopwatch",Equipment: Equipment{Id: "system",Label:"System", Manufacturer:"System", Model:"Builtin"},Color: ""}

		return c,nil
	}

	err:=db.One("Id",id,&c)

	if(err!=nil){
		return c,err
	}

	return c,nil

}

func getCommands() ([]Command){

	var Commands []Command

	db.All(&Commands)

	c := Command{Id:"delay",Label: "Delay",Icon: "fa-stopwatch",Equipment: Equipment{Id: "system",Label:"System", Manufacturer:"System", Model:"Builtin"},Color: ""}

	Commands = append(Commands,c)

	return Commands
}
//this method expects the path calculated without extensions so add it and then search for the path
func getCommandByPath(pth string) (Command,error){

	var c Command

	if(pth=="delay"){

		return Command{"delay","Delay","fa-stopwatch",Equipment{Id: "system",Label:"System", Manufacturer:"System", Model:"Builtin"},"","delay.txt"},nil

	}


	if(strings.Contains(pth,"@")) {

		pth = filepath.FromSlash(strings.Replace(pth, "@", "/", -1))

	}else{
		//no @ means we are in legacy mode
		pth = fmt.Sprintf("%s/%s","legacy_mode",pth)

	}

    npth := fmt.Sprintf("%s.txt",pth)

    log.Println(npth)

	err:=db.One("Path",npth,&c)

	if(err!=nil){
		return c,err
	}

	return c,nil

}

func getCommandsByPath() []string{

	var files []string

	fp := filepath.Join(configPath,"commands")

	log.Println(fp)

	err := filepath.Walk(fp, func(path string, info os.FileInfo, err error) error {

		if path == fp {
			return nil
		}

		path = strings.Replace(path, filepath.FromSlash(fp+"/cmd_"), "", -1)
		path = strings.Replace(path, filepath.FromSlash("commands/cmd_"), "", -1)

		log.Println(path)

		parts := strings.Split(path, ".")

		if parts[0] == "commands/" {
			return nil
		}

		files = append(files, parts[0])
		return nil
	})

	if(err!=nil){
		log.Println(err)
	}

	return files


}

func (c Command) ToCmd() (string){

   //magic bit for delay

	if(c.Id=="delay"){

		return "delay"
	}


	re := regexp.MustCompile(`\W`)
	lab := strings.Replace(c.Label," ","_",-1)
	lab = fmt.Sprintf("%s",re.ReplaceAll([]byte(lab),[]byte("")))


	if(len(c.Path)>0){

		parts := strings.Split(c.Path,"/")

		if(len(parts)>1){

			cmd := parts[1]

			lab = strings.TrimSuffix(cmd, filepath.Ext(cmd))

		}

	}

	pth := ""

	if(strings.ToLower(c.Equipment.Id)=="legacy_mode"){

		pth = fmt.Sprintf("%s", strings.ToLower(lab))

	}else {

		pth = fmt.Sprintf("%s@%s", strings.ToLower(c.Equipment.Id), strings.ToLower(lab))

	}


    return pth

}

func (c *Command) ToFullPath() (string){

	re := regexp.MustCompile(`\W`)
	lab := strings.Replace(c.Label," ","_",-1)
	lab = fmt.Sprintf("%s",re.ReplaceAll([]byte(lab),[]byte("")))
	fp := filepath.Join(configPath,"commands",strings.ToLower(c.Equipment.Id))

	return fp

}

func (c *Command) ToFullFile() (string){

	if(len(c.Path)>1){
		return filepath.Join(configPath,"commands",c.Path)
	}

	re := regexp.MustCompile(`\W`)
	lab := strings.Replace(c.Label," ","_",-1)
	lab = fmt.Sprintf("%s",re.ReplaceAll([]byte(lab),[]byte("")))
	top := filepath.Join(configPath,"commands",strings.ToLower(c.Equipment.Id))
	fp := filepath.Join(top,fmt.Sprintf("%s.txt",strings.ToLower(lab)))

    return fp

}

func (c *Command) GenerateId() {


	if(len(c.Id)<1){


		si := int(adler32.Checksum([]byte(fmt.Sprintf("%s%s",c.Label,c.Equipment.Id))))
		hd := hashids.NewData()
		hd.Salt = "irisonlyok"
		hd.MinLength = 10
		h, _ := hashids.NewWithData(hd)
		e, _ := h.Encode([]int{si})
		c.Id = e

	}

}

func (c *Command) ToShortPath()(string){

	re := regexp.MustCompile(`\W`)
	lab := strings.Replace(c.Label," ","_",-1)
	lab = fmt.Sprintf("%s",re.ReplaceAll([]byte(lab),[]byte("")))

	//this is used over filepath so that we always save a / in the db, this way we don't have unexpected issues if you move from windows to linu

	pth := strings.Join([]string{strings.ToLower(c.Equipment.Id),fmt.Sprintf("%s.txt",strings.ToLower(lab))},"/")

	return pth

}

func (c *Command) EquipmentRefresh() (error){


    c.Equipment = c.Equipment.Refresh()

	return c.Save()
}

func (c *Command) Remove() (error){


	return db.DeleteStruct(c)
}

func (c *Command) Save() (error){

	c.GenerateId()

	return db.Save(c)
}
