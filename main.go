package main

import (
	"arrowcloudapi/mongo"
	_ "arrowcloudapi/routers"
	"arrowcloudapi/utils"
	"errors"
	"fmt"
	"os"
	"strings"

	"arrowcloudapi/service/swarm"
	"arrowcloudapi/service/swarm/docker"

	"github.com/astaxie/beego"
	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/pkg/term"
	// "github.com/spf13/pflag"
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

	managerIP, err := findSwarmManager()
	if err != nil {
		panic(err)
	}
	log.Infof("Found swarm manager: %s", managerIP)

	// managerIP := "jin-onpremises"

	os.Setenv("DOCKER_HOST", "tcp://"+managerIP+":2376")
	os.Setenv("DOCKER_CERT_PATH", "/etc/docker-certs")

	err = swarm.Initialize()
	if err != nil {
		panic(err)
	}

	_, err = docker.ListNodes()

	dbOpts := map[string]string{
		"hostname": beego.AppConfig.String("MONGO_HOSTS"),
		"port":     beego.AppConfig.String("MONGO_PORT"),
		"rsname":   beego.AppConfig.String("MONGO_RSNAME"),
		"dbname":   beego.AppConfig.String("MONGO_DBNAME"),
		"username": beego.AppConfig.String("MONGO_USERNAME"),
		"password": beego.AppConfig.String("MONGO_PASSWORD"),
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

/*
	result of "docker info"
	{
	  "ID": "XOW3:6EEW:4DJ3:CSZJ:44PY:LIAS:UGXH:C7LZ:Z5JF:TBAV:FDBG:EU6D",
	  "Containers": 4,
	  "ContainersRunning": 0,
	  "ContainersPaused": 0,
	  "ContainersStopped": 4,
	  "Images": 1173,
	  "Driver": "overlay2",
	  "DriverStatus": [
	    [
	      "Backing Filesystem",
	      "extfs"
	    ],
	    [
	      "Supports d_type",
	      "true"
	    ],
	    [
	      "Native Overlay Diff",
	      "true"
	    ]
	  ],
	  "SystemStatus": null,
	  "Plugins": {
	    "Volume": [
	      "local"
	    ],
	    "Network": [
	      "bridge",
	      "host",
	      "macvlan",
	      "null",
	      "overlay"
	    ],
	    "Authorization": null,
	    "Log": null
	  },
	  "MemoryLimit": true,
	  "SwapLimit": false,
	  "KernelMemory": true,
	  "CpuCfsPeriod": true,
	  "CpuCfsQuota": true,
	  "CPUShares": true,
	  "CPUSet": true,
	  "IPv4Forwarding": true,
	  "BridgeNfIptables": true,
	  "BridgeNfIp6tables": true,
	  "Debug": false,
	  "NFd": 17,
	  "OomKillDisable": true,
	  "NGoroutines": 24,
	  "SystemTime": "2017-10-12T05:12:17.270537733Z",
	  "LoggingDriver": "json-file",
	  "CgroupDriver": "cgroupfs",
	  "NEventsListener": 0,
	  "KernelVersion": "4.4.0-78-generic",
	  "OperatingSystem": "Ubuntu 16.04.2 LTS",
	  "OSType": "linux",
	  "Architecture": "x86_64",
	  "IndexServerAddress": "https://index.docker.io/v1/",
	  "RegistryConfig": {
	    "AllowNondistributableArtifactsCIDRs": null,
	    "AllowNondistributableArtifactsHostnames": null,
	    "InsecureRegistryCIDRs": [
	      "127.0.0.0/8"
	    ],
	    "IndexConfigs": {
	      "docker.io": {
	        "Name": "docker.io",
	        "Mirrors": null,
	        "Secure": true,
	        "Official": true
	      },
	      "registry.cloudapp.jin.com": {
	        "Name": "registry.cloudapp.jin.com",
	        "Mirrors": [],
	        "Secure": false,
	        "Official": false
	      }
	    },
	    "Mirrors": []
	  },
	  "NCPU": 2,
	  "MemTotal": 7841349632,
	  "GenericResources": null,
	  "DockerRootDir": "/var/lib/docker",
	  "HttpProxy": "",
	  "HttpsProxy": "",
	  "NoProxy": "",
	  "Name": "ip-10-187-138-122",
	  "Labels": [
	    "HOST_IP=172.30.0.247",
	    "SSH_HOST_IP=13.112.198.7"
	  ],
	  "ExperimentalBuild": false,
	  "ServerVersion": "1.13.0",
	  "ClusterStore": "",
	  "ClusterAdvertise": "",
	  "Runtimes": {
	    "runc": {
	      "path": "docker-runc"
	    }
	  },
	  "DefaultRuntime": "runc",

	  "Swarm": {
	    "NodeID": "",
	    "NodeAddr": "",
	    "LocalNodeState": "inactive",
	    "ControlAvailable": false,
	    "Error": "",
	    "RemoteManagers": null,
	    "Cluster": {
	      "ID": "",
	      "Version": {},
	      "CreatedAt": "0001-01-01T00:00:00Z",
	      "UpdatedAt": "0001-01-01T00:00:00Z",
	      "Spec": {
	        "Labels": null,
	        "Orchestration": {},
	        "Raft": {
	          "ElectionTick": 0,
	          "HeartbeatTick": 0
	        },
	        "Dispatcher": {},
	        "CAConfig": {},
	        "TaskDefaults": {},
	        "EncryptionConfig": {
	          "AutoLockManagers": false
	        }
	      },
	      "TLSInfo": {},
	      "RootRotationInProgress": false
	    }
	  },
	  "LiveRestoreEnabled": false,
	  "Isolation": "",
	  "InitBinary": "docker-init",
	  "ContainerdCommit": {
	    "ID": "03e5862ec0d8d3b3f750e19fca3ee367e13c090e",
	    "Expected": "03e5862ec0d8d3b3f750e19fca3ee367e13c090e"
	  },
	  "RuncCommit": {
	    "ID": "2f7393a47307a16f8cee44a37b262e8b81021e3e",
	    "Expected": "2f7393a47307a16f8cee44a37b262e8b81021e3e"
	  },
	  "InitCommit": {
	    "ID": "949e6fa",
	    "Expected": "949e6fa"
	  },
	  "SecurityOptions": [
	    "name=apparmor",
	    "name=seccomp,profile=default"
	  ]
	}
*/
func findSwarmManager() (managerIP string, err error) {

	// https://github.com/docker/cli/blob/master/cmd/docker/docker.go#L165
	// https://github.com/docker/cli/blob/master/cli/command/cli.go#L180
	stdin, stdout, stderr := term.StdStreams()
	dockerCli := command.NewDockerCli(stdin, stdout, stderr)

	opts := cliflags.NewClientOptions()

	// When it tried to connect to the docker daemon via ip:port there was an error regarding incorrect protocol.
	// (http used for https reponse). The following code was for fixing it but it's not done.
	// flags := pflag.CommandLine
	// opts.Common.InstallFlags(flags)

	dockerCli.Initialize(opts)

	dockerInfo, err := dockerCli.Client().Info(context.Background())
	if err != nil {
		return
	}

	utils.PrettyPrint(dockerInfo)

	if dockerInfo.Swarm.LocalNodeState == "inactive" {
		errMsg := "Current node is not in a swarm."
		log.Warning(errMsg)
		err = errors.New(errMsg)
		return
	}

	managers := dockerInfo.Swarm.RemoteManagers
	if managers == nil || len(managers) == 0 {
		errMsg := "no available swarm manager found"
		log.Error(errMsg)
		err = errors.New(errMsg)
		return
	}

	// process the result
	// [{"NodeID":"z9j7jebvibi4rdnfdtmi2arod","Addr":"10.35.159.51:2377"},{"NodeID":"wxdz6sgkwq4pxc3b7ws8rqjwo","Addr":"10.35.149.229:2377"},{"NodeID":"rttfonyon3piqybyrbuixakjt","Addr":"10.29.42.57:2377"}]
	managerIPs := []string{}
	for _, manager := range managers {
		result := strings.Split(manager.Addr, ":")
		ip := result[0]
		managerIPs = append(managerIPs, ip)
	}
	availableManagerIP := managerIPs[0]

	if availableManagerIP != "" {
		return availableManagerIP, nil
	} else {
		err = errors.New(fmt.Sprintf("failed to find a swarm manager. managers: %v", managerIPs))
		return
	}

	// if !info.Swarm.ControlAvailable {
	// 	return errors.New("this node is not a swarm manager. Use \"docker swarm init\" or \"docker swarm join\" to connect this node to swarm and try again")
	// }
}
