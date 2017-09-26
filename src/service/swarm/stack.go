package swarm

import (
	"service/swarm/docker"
	"utils/log"
)

// List stacks by running "docker stack ls" command
func ListStacks() (output string, err error) {

	output, err = execStackCommand("ls")

	if err != nil {
		log.Errorf("Failed to exec 'docker stack ls'. %v", err)
	}

	return
}

// List stacks by calling docker daemon API
func ListStacksFromAPI() (map[string]int, error) {

	return docker.ListStacks()
}

// Deploy a stack by calling "docker stack deploy" command
func DeployStack(stackName, composeFile string) (output string, err error) {

	output, err = execStackCommand("deploy", "-c", composeFile, stackName)

	if err != nil {
		log.Errorf("Failed to exec 'docker stack deploy with compose file %s'. %v", composeFile, err)
	}

	return
}

func CheckServices(stackName string) (output string, err error) {

	output, err = execStackCommand("services", stackName)

	if err != nil {
		log.Errorf("Failed to exec 'docker stack services'. %v", err)
	}

	return

}

func GetServiceLog(stackName, serviceName string) (output string, err error) {

	output, err = execServiceCommand("logs", stackName+"_"+serviceName)

	if err != nil {
		log.Errorf("Failed to exec 'docker service logs'. %v", err)
	}

	return
}

func RemoveStack(stackName string) (output string, err error) {

	output, err = execStackCommand("rm", stackName)

	if err != nil {
		log.Errorf("Failed to exec 'docker stack rm'. %v", err)
	}

	return
}
