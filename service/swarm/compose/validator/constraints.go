package validator

import (
	"arrowcloudapi/models"
	"arrowcloudapi/utils/log"
	"errors"
	"fmt"

	composetypes "github.com/docker/cli/cli/compose/types"
)

func init() {
	// compose.RegisterValidator(&ConstraintsValidator{})
}

type ConstraintsValidator struct {
}

func (cv *ConstraintsValidator) Name() string {
	return "ConstraintsValidator"
}

/*
  verify constraints
    * should not allow any constraints for now
    * how about db service?
        * use a special constraint to identify db services
*/
func (cv *ConstraintsValidator) Validate(stack models.Stack, stackConfig *composetypes.Config, yamlMap *map[string]interface{}) []error {

	log.Debug("constraints validator is about to validate...")

	errs := []error{}

	for _, service := range stackConfig.Services {

		constraints := service.Deploy.Placement.Constraints

		if constraints != nil && len(constraints) > 0 {
			errMsg := fmt.Sprintf("The service %s has placement constraints. It's not supported.", service.Name)
			errs = append(errs, errors.New(errMsg))
		}

	}

	return errs
}
