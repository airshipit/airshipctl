package isogen

import (
	"github.com/spf13/cobra"

	"sigs.k8s.io/yaml"

	"opendev.org/airship/airshipctl/pkg/environment"
)

const (
	// TODO this should be part of a airshipctl config
	EphemeralClusterAnnotation = "airshipit.org/clustertype=ephemeral"
)

// Settings settings for isogen command
type Settings struct {
	*environment.AirshipCTLSettings

	// Configuration file (YAML-formatted) path for ISO builder container.
	IsogenConfigFile string
}

// InitFlags adds falgs and their default settings for isogen command
func (i *Settings) InitFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVarP(&i.IsogenConfigFile, "config", "c", "", "Configuration file path for ISO builder container.")
}

// Config ISO builder container configuration
type Config struct {
	// Configuration parameters for container
	Container Container `json:"container,omitempty"`
	// Configuration parameters for ISO builder
	Builder Builder `json:"builder,omitempty"`
}

// Container parameters
type Container struct {
	// Container volume directory binding.
	Volume string `json:"volume,omitempty"`
	// ISO generator container image URL
	Image string `json:"image,omitempty"`
	// Container Runtime Interface driver
	ContainerRuntime string `json:"containerRuntime,omitempty"`
}

// Builder parameters
type Builder struct {
	// Cloud Init user-data file name placed to the container volume root
	UserDataFileName string `json:"userDataFileName,omitempty"`
	// Cloud Init network-config file name placed to the container volume root
	NetworkConfigFileName string `json:"networkConfigFileName,omitempty"`
	// File name for output metadata
	OutputMetadataFileName string `json:"outputMetadataFileName,omitempty"`
}

// ToYAML serializes confid to YAML
func (c *Config) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}
