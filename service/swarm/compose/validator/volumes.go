package validator

import (
	"arrowcloudapi/service/swarm/compose"
	"arrowcloudapi/utils"
	"arrowcloudapi/utils/log"
	"errors"
	"fmt"

	composetypes "github.com/docker/cli/cli/compose/types"
)

func init() {
	compose.RegisterValidator(&VolumesValidator{})
}

type VolumesValidator struct {
}

func (vv *VolumesValidator) Name() string {
	return "VolumesValidator"
}

/*
 * a stack should have its own network
 * all services should be on the network
 */
func (vv *VolumesValidator) Validate(stackConfig *composetypes.Config, yamlMap *map[string]interface{}) []error {

	log.Debug("volumes validator is about to validate...")

	errs := []error{}

	for _, service := range stackConfig.Services {

		/* service level volumes
			ServiceVolumeConfig struct: https://github.com/docker/cli/blob/master/cli/compose/types/types.go#L257
			{
			  "Type": "bind",
			  "Source": "/db-data",
			  "Target": "/var/lib/postgresql/data",
			  "ReadOnly": false,
			  "Consistency": "",
			  "Bind": null,
			  "Volume": null
			}

			YAML
		   "volumes": [
		     "db-data:/var/lib/postgresql/data"
		   ]
		*/

		if service.Volumes == nil {
			continue
		}

		for _, volumeConfig := range service.Volumes {
			if volumeConfig.Type == "bind" {
				errMsg := fmt.Sprintf("Bind mount is not allowed. service: %s, source: %s, target: %s", service.Name, volumeConfig.Source, volumeConfig.Target)
				errs = append(errs, errors.New(errMsg))
				log.Error(errMsg)
			}
		}

	}

	/* top level volumes
	VolumeConfig struct: https://github.com/docker/cli/blob/master/cli/compose/types/types.go#L321
	{
	  "db-data": {
	    "Name": "",
	    "Driver": "local",
	    "DriverOpts": {
	      "device": ":/appc_data/dbstore",
	      "o": "addr=10.173.145.82,rw",
	      "type": "nfs"
	    },
	    "External": {
	      "Name": "",
	      "External": false
	    },
	    "Labels": null
	  }
	}

	{
	  "db-data": {
	    "Name": "",
	    "Driver": "",
	    "DriverOpts": null,
	    "External": {
	      "Name": "",
	      "External": false
	    },
	    "Labels": null
	  }
	}

	YAML
	"volumes": {
	  "db-data": null
	}
	*/

	utils.PrettyPrint(stackConfig.Volumes)

	for name, volumeConfig := range stackConfig.Volumes {
		if volumeConfig.External.External {
			errMsg := fmt.Sprintf("External volume is not allowed. volume name: %s", name)
			errs = append(errs, errors.New(errMsg))
			log.Error(errMsg)
		}

		if volumeConfig.Driver != "" && volumeConfig.Driver != "local" {
			errMsg := fmt.Sprintf("Volume driver %s is not supported.", volumeConfig.Driver)
			errs = append(errs, errors.New(errMsg))
			log.Error(errMsg)
		}

		if volumeConfig.DriverOpts != nil {
			errMsg := fmt.Sprintf("Volume driver_opts %v is not allowed.", volumeConfig.DriverOpts)
			errs = append(errs, errors.New(errMsg))
			log.Error(errMsg)
		}

		/*
		  "volumes": {
		    "db-data": {
		      "driver": "local",
		      "driver_opts": {
		        "device": ":/appc_data/dbstore",
		        "o": "addr=10.173.145.82,rw",
		        "type": "nfs"
		      }
		    }
		  }

		  The code below updates a stack's volume configuration to use NFS as storage.
		*/

		if volumeConfig.Driver == "" {
			volumesMap := (*yamlMap)["volumes"].(map[string]interface{})
			var volumeConfigMap map[string]interface{}
			if volumesMap[name] != nil {
				volumeConfigMap = volumesMap[name].(map[string]interface{})
			}
			if volumeConfigMap == nil {
				volumeConfigMap = map[string]interface{}{}
				volumesMap[name] = volumeConfigMap
			}

			volumeConfigMap["driver"] = "local"
			volumeConfigMap["driver_opts"] = map[string]interface{}{
				"device": ":/appc_data/dbstore",
				"o":      "addr=10.173.145.82,rw",
				"type":   "nfs",
			}
		}

	}

	return errs
}
