package validator

import (
	"arrowcloudapi/models"
	"arrowcloudapi/utils"
	"arrowcloudapi/utils/log"
	"fmt"
	"os"

	composetypes "github.com/docker/cli/cli/compose/types"
)

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

	servicesConfig := (*yamlMap)["services"].(map[string]interface{})

	for _, service := range stackConfig.Services {

		if service.Ports == nil {
			log.Debugf("The service %s does not expose ports.", service.Name)
			continue
		}

		// according to docker-flow-proxy the service does not need to expose any ports,
		// but we need to add labels for docker-flow-proxy to discover it.
		serviceConfig := servicesConfig[service.Name].(map[string]interface{})
		delete(serviceConfig, "ports")

		// The service needs to be attached to proxy_network so that haproxy can access it
		addServiceNetwork(service.Name, &serviceConfig, "proxy_network")
		serviceDomain := service.Name + "." + stack.ID + ".stack." + os.Getenv("ARROWCLOUD_DOMAIN")
		addServiceLabel(service.Name, &serviceConfig, "com.df.notify=true")
		addServiceLabel(service.Name, &serviceConfig, "com.df.distribute=true")
		addServiceLabel(service.Name, &serviceConfig, "com.df.serviceDomain="+serviceDomain)
		// https://github.com/vfarcic/docker-flow-proxy/issues/256
		// https://github.com/vfarcic/docker-flow-proxy/issues/140
		// (not related but useful: https://github.com/vfarcic/docker-flow-proxy/issues/287)
		// (not related but useful: https://github.com/vfarcic/docker-flow-proxy/issues/107)
		// (not related but useful: https://github.com/vfarcic/docker-flow-proxy/issues/293)
		addServiceLabel(service.Name, &serviceConfig, "com.df.serviceDomainMatchAll=true")

		// composetypes.ServicePortConfig
		log.Debugf("The service %s's ports config:", service.Name)
		for _, portConfig := range service.Ports {
			utils.PrettyPrint(portConfig)

			// http://proxy.dockerflow.com/usage/#reconfigure
			addServiceLabel(service.Name, &serviceConfig, "com.df.port="+fmt.Sprintf("%v", portConfig.Target))
		}
	}

	return errs
}
