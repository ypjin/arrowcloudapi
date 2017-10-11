package main

import (
	"arrowcloudapi/mongo"
	_ "arrowcloudapi/routers"
	"arrowcloudapi/utils"
	"os"

	"arrowcloudapi/service/swarm"
	"arrowcloudapi/service/swarm/docker"

	"github.com/astaxie/beego"
	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/pkg/term"
	"github.com/spf13/pflag"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2/bson"

	_ "arrowcloudapi/beego_ext"
	_ "arrowcloudapi/service/swarm/compose/validator"
	"arrowcloudapi/utils/log"
)

func main() {

	log.SetLevel(log.DebugLevel)

	// https://beego.me/docs/mvc/controller/config.md
	// https://beego.me/docs/mvc/controller/session.md
	beego.BConfig.WebConfig.Session.SessionOn = true
	beego.BConfig.WebConfig.Session.SessionProvider = "mongo"
	beego.BConfig.WebConfig.Session.SessionProviderConfig = ""
	beego.BConfig.WebConfig.Session.SessionName = "connect.sid"
	beego.BConfig.WebConfig.Session.SessionSecret = "stratus"
	beego.BConfig.WebConfig.Session.SessionGCMaxLifetime = 3600
	beego.BConfig.WebConfig.Session.SessionCookieLifeTime = 3600
	beego.BConfig.WebConfig.Session.SessionDomain = ""

	beego.AddTemplateExt("htm")

	// dao.InitDatabase()

	err := findSwarmManager()
	if err != nil {
		panic(err)
	}

	os.Setenv("DOCKER_HOST", "tcp://jin-onpremises:2376")
	os.Setenv("DOCKER_CERT_PATH", "/Users/yjin/onpremises-test")

	err = swarm.Initialize()
	if err != nil {
		panic(err)
	}

	_, err = docker.ListNodes()

	dbOpts := map[string]string{
		"hostname": "176.34.6.8",
		"port":     "60001",
		"dbname":   "arrowcloud",
		"username": "appcelerator",
		"password": "a5rSFu8RWDAP0jfc",
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

func findSwarmManager() error {

	// https://github.com/docker/cli/blob/master/cmd/docker/docker.go#L165
	// https://github.com/docker/cli/blob/master/cli/command/cli.go#L180
	stdin, stdout, stderr := term.StdStreams()
	dockerCli := command.NewDockerCli(stdin, stdout, stderr)

	opts := cliflags.NewClientOptions()

	flags := pflag.CommandLine
	opts.Common.InstallFlags(flags)

	dockerCli.Initialize(opts)

	if dockerCli == nil {
		log.Error("dockerCli is nil")
		return nil
	}

	if dockerCli.Client() == nil {
		log.Error("dockerCli.Client is nil")
		return nil
	}

	info, err := dockerCli.Client().Info(context.Background())
	if err != nil {
		return err
	}

	utils.PrettyPrint(info)

	// if !info.Swarm.ControlAvailable {
	// 	return errors.New("this node is not a swarm manager. Use \"docker swarm init\" or \"docker swarm join\" to connect this node to swarm and try again")
	// }

	return nil

}
