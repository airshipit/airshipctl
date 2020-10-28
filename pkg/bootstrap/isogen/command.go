/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package isogen

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/bootstrap/cloudinit"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/util"
)

const (
	builderConfigFileName  = "builder-conf.yaml"
	outputFileNameDefault  = "ephemerial.iso"
	userDataFileName       = "user-data"
	networkConfigFileName  = "network-data"
	outputMetadataFileName = "output-metadata.yaml"
)

// BootstrapIsoOptions are used to generate bootstrap ISO
type BootstrapIsoOptions struct {
	DocBundle document.Bundle
	Builder   container.Container
	Doc       document.Document
	Cfg       *v1alpha1.IsoConfiguration

	// optional fields for verbose output
	Writer io.Writer
}

func VerifyInputs(cfg *v1alpha1.IsoConfiguration) error {
	if cfg.IsoContainer.Volume == "" {
		return config.ErrMissingConfig{
			What: "Must specify volume bind for ISO builder container",
		}
	}

	vols := strings.Split(cfg.IsoContainer.Volume, ":")
	switch {
	case len(vols) == 1:
		cfg.IsoContainer.Volume = fmt.Sprintf("%s:%s", vols[0], vols[0])
	case len(vols) > 2:
		return config.ErrInvalidConfig{
			What: "Bad container volume format. Use hostPath:contPath",
		}
	}

	if cfg.Isogen.OutputFileName == "" {
		log.Debugf("No outputFileName provided to Isogen. Using default: %s", outputFileNameDefault)
		cfg.Isogen.OutputFileName = outputFileNameDefault
	}

	return nil
}

func getIsoContainerCfg(
	cfg *v1alpha1.IsoConfiguration,
	builderCfgYaml []byte,
	userData []byte,
	netConf []byte,
) map[string][]byte {
	hostVol := strings.Split(cfg.IsoContainer.Volume, ":")[0]

	fls := make(map[string][]byte)
	fls[filepath.Join(hostVol, userDataFileName)] = userData
	fls[filepath.Join(hostVol, networkConfigFileName)] = netConf
	fls[filepath.Join(hostVol, builderConfigFileName)] = builderCfgYaml
	return fls
}

// CreateBootstrapIso prepares and runs appropriate container to create a bootstrap ISO
func (opts BootstrapIsoOptions) CreateBootstrapIso() error {
	cntVol := strings.Split(opts.Cfg.IsoContainer.Volume, ":")[1]
	log.Print("Creating cloud-init for ephemeral K8s")
	userData, netConf, err := cloudinit.GetCloudData(
		opts.DocBundle,
		opts.Cfg.Isogen.UserDataSelector,
		opts.Cfg.Isogen.UserDataKey,
		opts.Cfg.Isogen.NetworkConfigSelector,
		opts.Cfg.Isogen.NetworkConfigKey,
	)
	if err != nil {
		return err
	}

	builderCfgYaml, err := opts.Doc.AsYAML()
	if err != nil {
		return err
	}

	fls := getIsoContainerCfg(opts.Cfg, builderCfgYaml, userData, netConf)
	if err = util.WriteFiles(fls, 0600); err != nil {
		return err
	}

	vols := []string{opts.Cfg.IsoContainer.Volume}
	builderCfgLocation := filepath.Join(cntVol, builderConfigFileName)
	log.Printf("Running default container command. Mounted dir: %s", vols)

	envVars := []string{
		fmt.Sprintf("IMAGE_TYPE=iso"),
		fmt.Sprintf("BUILDER_CONFIG=%s", builderCfgLocation),
		fmt.Sprintf("USER_DATA_FILE=%s", userDataFileName),
		fmt.Sprintf("NET_CONFIG_FILE=%s", networkConfigFileName),
		fmt.Sprintf("OUTPUT_FILE_NAME=%s", opts.Cfg.Isogen.OutputFileName),
		fmt.Sprintf("OUTPUT_METADATA_FILE_NAME=%s", outputMetadataFileName),
		fmt.Sprintf("http_proxy=%s", os.Getenv("http_proxy")),
		fmt.Sprintf("https_proxy=%s", os.Getenv("https_proxy")),
		fmt.Sprintf("HTTP_PROXY=%s", os.Getenv("HTTP_PROXY")),
		fmt.Sprintf("HTTPS_PROXY=%s", os.Getenv("HTTPS_PROXY")),
		fmt.Sprintf("no_proxy=%s", os.Getenv("no_proxy")),
		fmt.Sprintf("NO_PROXY=%s", os.Getenv("NO_PROXY")),
	}

	err = opts.Builder.RunCommand([]string{}, nil, vols, envVars)
	if err != nil {
		return err
	}

	log.Print("ISO generation is in progress. The whole process could take up to several minutes, please wait...")

	if log.DebugEnabled() {
		var cLogs io.ReadCloser
		cLogs, err = opts.Builder.GetContainerLogs()
		if err != nil {
			log.Printf("failed to read container logs %s", err)
		} else {
			log.Print("start reading container logs")
			if _, err = io.Copy(opts.Writer, cLogs); err != nil {
				log.Debugf("failed to write container logs to log output %s", err)
			}
			log.Print("got EOF from container logs")
		}
	}

	if err = opts.Builder.WaitUntilFinished(); err != nil {
		return err
	}

	log.Print("ISO successfully built.")
	if !log.DebugEnabled() {
		log.Print("Removing container.")
		return opts.Builder.RmContainer()
	}

	log.Debugf("Debug flag is set. Container %s stopped but not deleted.", opts.Builder.GetID())
	return nil
}
