package validator

import (
	"arrowcloudapi/models"
	"arrowcloudapi/service/swarm/compose"
	"arrowcloudapi/utils"
	"arrowcloudapi/utils/log"

	composetypes "github.com/docker/cli/cli/compose/types"
)

func init() {
	compose.RegisterValidator(&PortsValidator{})
}

type PortsValidator struct {
}

func (pv *PortsValidator) Name() string {
	return "PortsValidator"
}

/*
  Services in a stack do not need to publish ports for inter-communication.
  For the services which need to be accessed from outside there are two options.
  * publish ports to the ingress network
  * join another swarm-wide network which haproxy can access
  If we go with the second option there is no need to publish ports for any services, but we should still require
  user to define ports to publish in the compose file so that we can identify which services need to be accessed
  from haproxy. The compose file needs to be updated to remove port defintions.
  For the first option, the compose file needs to be updated to remove publish ports in port definitions. We could
  not use the publish ports user defined directly because there may be conficts with other services's ports.
*/
func (pv *PortsValidator) Validate(stack models.Stack, stackConfig *composetypes.Config, yamlMap *map[string]interface{}) []error {

	log.Debug("ports validator is about to validate...")

	errs := []error{}

	for _, service := range stackConfig.Services {

		if service.Ports == nil {
			log.Debugf("The service %s does not expose ports.", service.Name)
			continue
		}

		// composetypes.ServicePortConfig
		log.Debugf("The service %s's ports config:", service.Name)
		for _, portConfig := range service.Ports {
			utils.PrettyPrint(portConfig)
		}

	}

	return errs
}
