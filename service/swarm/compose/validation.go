package compose

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli/compose/loader"
	composetypes "github.com/docker/cli/cli/compose/types"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"

	"arrowcloudapi/utils/log"
)

var (
	validators = make(map[string]Validator)
)

type Validator interface {
	Validate(stackConfig *composetypes.Config, yamlMap *map[string]interface{}) error
	Name() string
}

func RegisterValidator(v Validator) {
	validators[v.Name()] = v
}

// https://github.com/docker/cli/blob/master/cli/command/stack/deploy.go#L23
func Validate(composefile string) (*composetypes.Config, error) {

	log.Debugf("about to verify compose file: %s", composefile)

	// var details composetypes.ConfigDetails
	// https://github.com/docker/cli/blob/master/cli/compose/types/types.go#L57
	configDetails, err := getConfigDetails(composefile)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	//utils.PrettyPrint(config)

	// https://github.com/docker/cli/tree/master/cli/compose/loader
	// map[string]interface{}
	configYaml := configDetails.ConfigFiles[0].Config

	// utils.PrettyPrint(configYaml)

	// run validators (transformers)
	for name, v := range validators {
		err = v.Validate(config, &configYaml)
		if err != nil {
			log.Errorf("validator %s error: %v", name, err)
			return nil, err
		}
	}

	var yamlBytes []byte
	yamlBytes, err = yaml.Marshal(configYaml)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile("/Users/yjin/aaa.yaml", yamlBytes, os.FileMode(0644))

	return config, err

}

// https://github.com/docker/cli/blob/master/cli/command/stack/deploy_composefile.go#L122
func getConfigDetails(composefile string) (composetypes.ConfigDetails, error) {
	var details composetypes.ConfigDetails

	absPath, err := filepath.Abs(composefile)
	if err != nil {
		return details, err
	}
	details.WorkingDir = filepath.Dir(absPath)

	configFile, err := getConfigFile(composefile)
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
