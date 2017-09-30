package swarm

import (
	"github.com/docker/cli/cli/compose/loader"
	composetypes "github.com/docker/cli/cli/compose/types"
)

func func_name() (*composetypes.Config, error) {

	var details composetypes.ConfigDetails

	config, err := loader.Load(details)

	return config, err

}
