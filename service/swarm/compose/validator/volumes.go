package validator

import (
	"arrowcloudapi/models"
	"arrowcloudapi/utils"
	"arrowcloudapi/utils/log"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	composetypes "github.com/docker/cli/cli/compose/types"
)

type VolumesValidator struct {
}

func (vv *VolumesValidator) Name() string {
	return "VolumesValidator"
}

/*
 * a stack should have its own network
 * all services should be on the network
 */
func (vv *VolumesValidator) Validate(stack *models.Stack, stackConfig *composetypes.Config, yamlMap *map[string]interface{}) []error {

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
	folderNames := []string{}

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

			folderName, err := createVolumeFolder(stack, name)
			if err != nil {
				errs = append(errs, err)
				return errs
			}
			folderNames = append(folderNames, folderName)

			volumeConfigMap["driver"] = "local"
			volumeConfigMap["driver_opts"] = map[string]interface{}{
				"device": ":/appc_data/stack-volume/" + folderName,
				"o":      "addr=" + os.Getenv("NFS_SERVER_IP") + ",rw",
				"type":   "nfs",
			}
		}

	}
	stack.VolumeFolders = strings.Join(folderNames, ",")

	return errs
}

func createVolumeFolder(stack *models.Stack, volumeName string) (folderName string, err error) {

	folderName = stack.ID + "_" + stack.Name + "_" + volumeName
	folderPath := path.Join("/volume_home", folderName)
	os.MkdirAll(folderPath, 0777)
	return
}
