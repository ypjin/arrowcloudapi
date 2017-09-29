package swarm

import (
	"arrowcloudapi/service/swarm/docker"

	"github.com/docker/docker/api/types/swarm"
)

func ListServices() ([]swarm.Service, error) {

	return docker.ListServices()

}
