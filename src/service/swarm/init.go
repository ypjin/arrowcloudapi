package swarm

import (
	"service/swarm/docker"
)

func Initialize() error {

	host := docker.GetHostSpec("54.168.32.62")

	return docker.InitDockerClient(host, "/Users/yjin/onpremises-test")
}
