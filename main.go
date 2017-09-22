package main

import (
	"mongo"
	_ "routers"

	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"

	"service/swarm"
	"service/swarm/docker"

	_ "beego_ext"
	"utils/log"
)

func main() {

	log.SetLevel(log.DebugLevel)

	// https://beego.me/docs/mvc/controller/config.md
	// https://beego.me/docs/mvc/controller/session.md
	beego.BConfig.WebConfig.Session.SessionOn = true
	beego.BConfig.WebConfig.Session.SessionProvider = "mongo"
	beego.BConfig.WebConfig.Session.SessionProviderConfig = ""
	beego.BConfig.WebConfig.Session.SessionName = "beegosessionID"
	beego.BConfig.WebConfig.Session.SessionGCMaxLifetime = 3600
	beego.BConfig.WebConfig.Session.SessionCookieLifeTime = 3600
	beego.BConfig.WebConfig.Session.SessionDomain = ""

	beego.AddTemplateExt("htm")

	// dao.InitDatabase()

	err := swarm.Initialize()
	if err != nil {
		panic(err)
	}

	_, err = docker.ListNodes()

	dbOpts := map[string]string{
		"hostname": "localhost",
		"port":     "27017",
		"dbname":   "arrowcloud",
	}

	err = mongo.Initialize(dbOpts)
	if err != nil {
		panic(err)
	}

	type Person struct {
		Name  string
		Phone string
	}

	// session := mongo.GetSession()
	// c := session.DB("arrowcloud").C("people")
	// err = c.Insert(
	// 	&Person{"Ale", "+55 53 8116 9639"},
	// 	&Person{"Cla", "+55 53 8402 8510"})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// result := Person{}
	// err = c.Find(bson.M{"name": "Ale"}).One(&result)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println("Phone:", result.Phone)

	re, err := mongo.FindOneDocument("arrowcloud:people", bson.M{"name": "Ale"})
	log.Infof("err: %v", err)
	log.Infof("persion: %v", re)

	beego.Run()
}
