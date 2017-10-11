package validator

import (
	"arrowcloudapi/service/swarm/compose"
	"arrowcloudapi/utils/log"
	"errors"
	"fmt"

	composetypes "github.com/docker/cli/cli/compose/types"
)

func init() {
	compose.RegisterValidator(&NetworksValidator{})
}

type NetworksValidator struct {
}

func (nv *NetworksValidator) Name() string {
	return "NetworksValidator"
}

/*
 * a stack should have its own network
 * all services should be on the network
 */
func (nv *NetworksValidator) Validate(stackConfig *composetypes.Config, yamlMap *map[string]interface{}) []error {

	log.Debug("networks validator is about to validate...")

	errs := []error{}

	if len(stackConfig.Networks) == 0 {
		log.Error("The stack does not have network definition.")
	}

	for _, service := range stackConfig.Services {

		if len(service.Networks) == 0 {
			errMsg := fmt.Sprintf("Service %s does not have network definition.", service.Name)
			errs = append(errs, errors.New(errMsg))
			log.Error(errMsg)
		}
	}

	return errs
}
