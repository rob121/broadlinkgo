package main

import(
	"runtime"
	"os"
	"strings"
	"path/filepath"
	"fmt"
	"log"
	"io"
)

func bootstrapOldCommands(){

	pth := getLegacyConfigPath()
	root:=filepath.Join(pth,"commands")

	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if(info.IsDir()){
			return nil
		}

		if (strings.Contains(path,"cmd_")) {

			files = append(files, path)
		}
		return nil
	})

	if err != nil {

	}

	e := addEquipment("Legacy","Legacy","Mode")

	errs := e.Save()

	fp := filepath.Join(configPath,"commands","legacy_mode")

	if(errs!=nil){

		log.Println(errs)
	}

	os.MkdirAll(fp, os.ModePerm)

	for _, file := range files {
		//create commands

		s1 := strings.Replace(filepath.Base(file),filepath.Ext(file),"",-1)
		s2 := strings.Replace(s1,"cmd_","",-1)

		label := strings.Replace(s2,"_"," ",-1)

		c:=addCommand(label,"",e,"")

		from := file
		to := filepath.Join(configPath,"commands",filepath.FromSlash(c.Path))

		err := CopyFile(from,to)

		if err != nil {
			fmt.Printf("Copy File %s failed %q\n", to,err)
		} else {
			fmt.Printf("Copy File  %s succeeded,upgrade completed\n",to)
		}

		c.Save()
	}
}

func getLegacyConfigPath() string{


	var cpath=""

	if ( runtime.GOOS == "windows") {

		cpath = os.Getenv("APPDATA")+"\\broadlinkgo\\"


	}else if(runtime.GOOS == "darwin"){

		cpath = os.Getenv("HOME")+"/broadlinkgo/"

	}else{

		cpath = "/etc/broadlinkgo/"
	}


	cmdpath = strings.Replace(cpath,"\\","/",-1)//we do this because we use filepath everywhere and need the same file path direction
	cmdpath = strings.Replace(cmdpath,"\"","",-1)

	slash := cmdpath[len(cmdpath)-1:]

	if(slash!="/"){

		cmdpath = cmdpath+"/"

	}

	return cmdpath


}


// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}