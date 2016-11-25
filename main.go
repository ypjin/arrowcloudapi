package main

import (
	_ "routers"

	"github.com/astaxie/beego"
)

func main() {

	beego.BConfig.WebConfig.Session.SessionOn = true

	beego.AddTemplateExt("htm")

	beego.Run()
}
