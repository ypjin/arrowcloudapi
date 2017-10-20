package compose

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/docker/cli/cli/compose/loader"
	composetypes "github.com/docker/cli/cli/compose/types"
	"github.com/pkg/errors"

	"arrowcloudapi/models"
	"arrowcloudapi/service/swarm/compose/validator"
	"arrowcloudapi/utils"
	"arrowcloudapi/utils/log"
)

var (
	validators []Validator
)

type Validator interface {
	Validate(stack models.Stack, stackConfig *composetypes.Config, yamlMap *map[string]interface{}) []error
	Name() string
}

func init() {
	validators = []Validator{
		// &validator.ConstraintsValidator{},
		&validator.NetworksValidator{},
		&validator.PortsValidator{},
		&validator.VolumesValidator{},
		&validator.LabelsValidator{},
	}
}

// func RegisterValidator(v Validator) {
// 	validators[v.Name()] = v
// }

// https://github.com/docker/cli/blob/master/cli/command/stack/deploy.go#L23
func Validate(stack *models.Stack, composeFile string) (*map[string]interface{}, []error) {

	log.Debugf("about to verify compose file: %s", composeFile)

	// var details composetypes.ConfigDetails
	// https://github.com/docker/cli/blob/master/cli/compose/types/types.go#L57
	configDetails, err := getConfigDetails(composeFile)
	if err != nil {
		return nil, []error{err}
	}

	// composetypes.Config
	// https://github.com/docker/cli/blob/master/cli/compose/types/types.go#L70
	/*
		type Config struct {
			Services []ServiceConfig
			Networks map[string]NetworkConfig
			Volumes  map[string]VolumeConfig
			Secrets  map[string]SecretConfig
			Configs  map[string]ConfigObjConfig
		}
	*/
	config, err := loader.Load(configDetails)
	if err != nil {
		return nil, []error{err}
	}

	//utils.PrettyPrint(config)

	// https://github.com/docker/cli/tree/master/cli/compose/loader
	// map[string]interface{}
	configYaml := configDetails.ConfigFiles[0].Config

	utils.PrettyPrint(configYaml)

	var yamlBytes []byte
	yamlBytes, err = yaml.Marshal(configYaml)
	if err != nil {
		return nil, []error{err}
	}
	stack.OriginalComposeFile = string(yamlBytes)

	// run validators (transformers)
	allValidationErrs := []error{}
	for _, v := range validators {
		errs := v.Validate(*stack, config, &configYaml)
		if errs != nil && len(errs) > 0 {
			allValidationErrs = append(allValidationErrs, errs...)
		}
	}

	if allValidationErrs != nil && len(allValidationErrs) > 0 {
		log.Errorf("validation errors: %v", allValidationErrs)
		return nil, allValidationErrs
	}

	utils.PrettyPrint(configYaml)

	return &configYaml, []error{}
}

// https://github.com/docker/cli/blob/master/cli/command/stack/deploy_composefile.go#L122
func getConfigDetails(composeFile string) (composetypes.ConfigDetails, error) {
	var details composetypes.ConfigDetails

	absPath, err := filepath.Abs(composeFile)
	if err != nil {
		return details, err
	}
	details.WorkingDir = filepath.Dir(absPath)

	configFile, err := getConfigFile(composeFile)
	if err != nil {
		return details, err
	}
	// TODO: support multiple files
	details.ConfigFiles = []composetypes.ConfigFile{*configFile}
	details.Environment, err = buildEnvironment(os.Environ())
	return details, err
}

// https://github.com/docker/cli/blob/master/cli/command/stack/deploy_composefile.go#L162
func getConfigFile(filename string) (*composetypes.ConfigFile, error) {
	var bytes []byte
	var err error

	bytes, err = ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config, err := loader.ParseYAML(bytes)
	if err != nil {
		return nil, err
	}

	return &composetypes.ConfigFile{
		Filename: filename,
		Config:   config,
	}, nil
}

// https://github.com/docker/cli/blob/master/cli/command/stack/deploy_composefile.go#L149
func buildEnvironment(env []string) (map[string]string, error) {
	result := make(map[string]string, len(env))
	for _, s := range env {
		// if value is empty, s is like "K=", not "K".
		if !strings.Contains(s, "=") {
			return result, errors.Errorf("unexpected environment %q", s)
		}
		kv := strings.SplitN(s, "=", 2)
		result[kv[0]] = kv[1]
	}
	return result, nil
}
