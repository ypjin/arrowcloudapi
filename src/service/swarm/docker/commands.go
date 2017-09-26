package docker

import (
	"context"
	"net/http"
	"path/filepath"

	"utils/log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/pkg/errors"
)

const (
	// LabelNamespace is the label used to track stack resources
	LabelNamespace = "com.docker.stack.namespace"
)

var (
	dockerClient *client.Client //client for swarm manager leader
)

func InitDockerClient(host string, dockerCertPath string) (err error) {

	if dockerClient != nil {
		return
	}

	log.Infof("Initialize docker client...")
	cli, err := connectToHost(host, dockerCertPath)
	dockerClient = cli

	if err != nil {
		return
	}

	_, err = InspectSwarm()
	return
}

func DestroyDockerClient() {
	if dockerClient != nil {
		dockerClient = nil
	}
}

func InspectSwarm() (swarm swarm.Swarm, err error) {
	return dockerClient.SwarmInspect(context.Background())
}

func ListNodes() (nodes []swarm.Node, err error) {

	options := types.NodeListOptions{}

	log.Infof("%v", dockerClient)

	return dockerClient.NodeList(context.Background(), options)
}

// https://github.com/docker/cli/blob/master/cli/command/stack/list.go
func ListStacks() (map[string]int, error) {

	options := types.ServiceListOptions{}
	options.Filters = getAllStacksFilter()

	services, err := dockerClient.ServiceList(context.Background(), options)
	if err != nil {
		return nil, err
	}

	m := make(map[string]int)

	for _, service := range services {
		labels := service.Spec.Labels
		name, ok := labels[LabelNamespace]
		if !ok {
			return nil, errors.Errorf("cannot get label %s for service %s",
				LabelNamespace, service.ID)
		}
		numServices, ok := m[name]
		if !ok {
			m[name] = 1
		} else {
			m[name] = numServices + 1
		}
	}

	return m, nil

}

// https://github.com/docker/cli/blob/master/cli/command/stack/common.go
func getStackFilter(namespace string) filters.Args {
	filter := filters.NewArgs()
	filter.Add("label", LabelNamespace+"="+namespace)
	return filter
}

func getAllStacksFilter() filters.Args {
	filter := filters.NewArgs()
	filter.Add("label", LabelNamespace)
	return filter
}

func ListServices() (services []swarm.Service, err error) {

	options := types.ServiceListOptions{}
	return dockerClient.ServiceList(context.Background(), options)
}

func RemoveService(serviceID string) error {
	return dockerClient.ServiceRemove(context.Background(), serviceID)
}

func CreateService(serviceSpec swarm.ServiceSpec) (types.ServiceCreateResponse, error) {
	options := types.ServiceCreateOptions{}
	return dockerClient.ServiceCreate(context.Background(), serviceSpec, options)
}

func ListTasks(serviceName string) (tasks []swarm.Task, err error) {

	filters := filters.NewArgs()
	filters.Add("service", serviceName)

	options := types.TaskListOptions{
		Filters: filters,
	}
	return dockerClient.TaskList(context.Background(), options)
}

func ListAllTasks() (tasks []swarm.Task, err error) {

	options := types.TaskListOptions{}
	return dockerClient.TaskList(context.Background(), options)
}

func ListContainers(host string, dockerCertPath string) (containers []types.Container, err error) {
	cli, err := connectToHost(host, dockerCertPath)
	if err != nil {
		return
	}
	options := types.ContainerListOptions{All: true}
	containers, err = cli.ContainerList(context.Background(), options)
	return
}

func DockerInfo(host string, dockerCertPath string) (info types.Info, err error) {
	cli, err := connectToHost(host, dockerCertPath)
	if err != nil {
		return
	}
	return cli.Info(context.Background())
}

func InspectNode(nodeID string) (res swarm.Node, err error) {
	res, _, err = dockerClient.NodeInspectWithRaw(context.Background(), nodeID)
	return
}

func InspectTask(taskID string) (res swarm.Task, err error) {
	res, _, err = dockerClient.TaskInspectWithRaw(context.Background(), taskID)
	return
}

func connectToHost(host string, dockerCertPath string) (cli *client.Client, err error) {

	cli, err = newSSLClient(host, "v1.22", dockerCertPath)
	return
}

func newSSLClient(host string, version string, dockerCertPath string) (*client.Client, error) {
	var httpClient *http.Client
	options := tlsconfig.Options{
		CAFile:             filepath.Join(dockerCertPath, "ca.pem"),
		CertFile:           filepath.Join(dockerCertPath, "cert.pem"),
		KeyFile:            filepath.Join(dockerCertPath, "key.pem"),
		InsecureSkipVerify: true,
	}
	tlsc, err := tlsconfig.Client(options)
	if err != nil {
		return nil, err
	}

	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsc,
		},
	}

	return client.NewClient(host, version, httpClient, nil)
}

func GetHostSpec(hostIP string) string {
	return "tcp://" + hostIP + ":2376"
}
