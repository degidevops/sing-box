package main

import (
	"os"
	"os/exec"
	_ "github.com/sagernet/gomobile/asset"
	"github.com/sagernet/sing-box/log"
)

func main() {
	findSDK()
	findMobile()
	command := exec.Command(
		goBinPath+"/gomobile", "bind",
		"-v",
		"-androidapi", "21",
		"-trimpath", "-ldflags=-s -w -buildid=",
		"-javapkg=io.nekohasekai",
		"-libname=box",
		"./experimental/libbox",
	)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		log.Fatal(err)
	}
	os.Rename("libbox.aar", "build/libbox.aar")
	os.Rename("libbox-sources.jar", "build/libbox-sources.jar")

}
