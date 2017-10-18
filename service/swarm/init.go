package swarm

import (
	"arrowcloudapi/service/swarm/docker"
	"os"
)

func Initialize() error {

	host := os.Getenv("DOCKER_HOST")
	hostSpec := docker.GetHostSpec(host)
	certPath := os.Getenv("DOCKER_CERT_PATH")

	return docker.InitDockerClient(hostSpec, certPath)
}
