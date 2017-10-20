package validator

import (
	"arrowcloudapi/models"
	"arrowcloudapi/utils"
	"arrowcloudapi/utils/log"
	"errors"
	"fmt"

	composetypes "github.com/docker/cli/cli/compose/types"
)

type NetworksValidator struct {
}

func (nv *NetworksValidator) Name() string {
	return "NetworksValidator"
}

/*
 * a stack should have its own network
 * all services should be on the network
 */
func (nv *NetworksValidator) Validate(stack models.Stack, stackConfig *composetypes.Config, yamlMap *map[string]interface{}) []error {

	log.Debug("networks validator is about to validate...")

	errs := []error{}

	if len(stackConfig.Networks) == 0 {
		log.Error("The stack does not have network definition.")
		errMsg := "A stack must include a network and attach all services to it."
		errs = append(errs, errors.New(errMsg))
	}

	for _, service := range stackConfig.Services {

		if len(service.Networks) == 0 {
			errMsg := fmt.Sprintf("Service %s does not have network definition.", service.Name)
			errs = append(errs, errors.New(errMsg))
			log.Error(errMsg)
		}
	}

	if len(errs) == 0 {
		addStackNetwork(yamlMap)
	}

	return errs
}

func addServiceNetwork(serviceName string, serviceConfig *map[string]interface{}, network string) {

	networkInf := (*serviceConfig)["networks"]

	var networks []interface{}
	if networkInf == nil {
		networks = []interface{}{}
	} else {
		networks = networkInf.([]interface{})
	}

	networks = append(networks, network)
	(*serviceConfig)["networks"] = networks

	log.Debugf("networks config of service %s", serviceName)
	utils.PrettyPrint(networks)
}

func addStackNetwork(composeMap *map[string]interface{}) {

	networks := (*composeMap)["networks"].(map[string]interface{})

	networks["proxy_network"] = map[string]interface{}{
		"external": true,
	}
}
