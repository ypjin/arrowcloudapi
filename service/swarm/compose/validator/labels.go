package validator

import (
	"arrowcloudapi/models"
	"arrowcloudapi/utils"
	"arrowcloudapi/utils/log"

	composetypes "github.com/docker/cli/cli/compose/types"
)

type LabelsValidator struct {
}

func (lv *LabelsValidator) Name() string {
	return "LabelsValidator"
}

/*
  verify labels
    * We are going to add com.axway.stack.id label here.
*/
func (lv *LabelsValidator) Validate(stack models.Stack, stackConfig *composetypes.Config, yamlMap *map[string]interface{}) []error {

	log.Debug("labels validator is about to validate...")

	customLabel := "com.axway.stack.id=" + stack.ID
	errs := []error{}

	for name, serviceInf := range (*yamlMap)["services"].(map[string]interface{}) {

		log.Debugf("add com.axway.stack.id label for service %s", name)

		service := serviceInf.(map[string]interface{})

		addServiceLabel(name, &service, customLabel)
	}

	return errs
}

func addServiceLabel(serviceName string, serviceConfig *map[string]interface{}, label string) {

	deployInf := (*serviceConfig)["deploy"]
	var deploy map[string]interface{}
	if deployInf == nil {
		deploy = map[string]interface{}{}
	} else {
		deploy = deployInf.(map[string]interface{})
	}
	(*serviceConfig)["deploy"] = deploy

	labelsInf := deploy["labels"]
	var labels []interface{}
	if labelsInf == nil {
		labels = []interface{}{}
	} else {
		labels = labelsInf.([]interface{})
	}

	labels = append(labels, label)
	deploy["labels"] = labels

	log.Debugf("deploy config of service %s", serviceName)
	utils.PrettyPrint(deploy)
}
