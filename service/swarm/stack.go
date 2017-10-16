package swarm

import (
	"arrowcloudapi/dao"
	"arrowcloudapi/models"
	"arrowcloudapi/service/swarm/compose"
	"arrowcloudapi/service/swarm/docker"
	"arrowcloudapi/utils/log"
	"errors"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
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
func ListStacksFromAPI(stackIds []string) (map[string]int, error) {

	return docker.ListStacks(stackIds)
}

// Deploy a stack by calling "docker stack deploy" command
func DeployStack(stack models.Stack, composeFile string) (output string, err error) {

	// *map[string]interface{}
	transformedConfigYaml, errs := compose.Validate(stack, composeFile)
	if errs != nil && len(errs) > 0 {
		log.Errorf("Failed to verify the compose file. %v", errs)

		errMsg := ""
		for _, e := range errs {
			errMsg += (e.Error() + "\n")
		}

		err = errors.New(errMsg)
		return
	}

	var yamlBytes []byte
	yamlBytes, err = yaml.Marshal(transformedConfigYaml)
	if err != nil {
		return
	}

	transformedComposeFile := "/Users/yjin/aaa.yaml"
	err = ioutil.WriteFile(transformedComposeFile, yamlBytes, os.FileMode(0644))
	if err != nil {
		return
	}

	output, err = execStackCommand("deploy", "-c", transformedComposeFile, stack.Name)

	if err != nil {
		log.Errorf("Failed to exec 'docker stack deploy with compose file %s'. %v", transformedComposeFile, err)
	}

	stack.ComposeFile = string(yamlBytes)
	_, err = dao.SaveStack(stack)
	if err != nil {
		log.Errorf("Failed to update stack compose file to db. %v", err)
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
