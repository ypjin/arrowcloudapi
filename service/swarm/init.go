package swarm

import (
	"arrowcloudapi/service/swarm/docker"
	"os"
)

func Initialize() error {

	host := os.Getenv("DOCKER_HOST")
	certPath := os.Getenv("DOCKER_CERT_PATH")

	return docker.InitDockerClient(host, certPath)
}
