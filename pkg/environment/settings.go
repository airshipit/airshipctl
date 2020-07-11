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

package environment

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	"k8s.io/client-go/tools/clientcmd"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/log"
)

// AirshipCTLSettings is a container for all of the settings needed by airshipctl
type AirshipCTLSettings struct {
	// Debug is used for verbose output
	Debug             bool
	AirshipConfigPath string
	KubeConfigPath    string
	Create            bool
	Config            *config.Config
}

// A singleton for the kustomize plugin path configuration
var pluginPath string
var pluginPathLock = &sync.Mutex{}

// InitFlags adds the default settings flags to cmd
func (a *AirshipCTLSettings) InitFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.BoolVar(
		&a.Debug,
		"debug",
		false,
		"enable verbose output")

	defaultAirshipConfigDir := filepath.Join(HomeEnvVar, config.AirshipConfigDir)

	defaultAirshipConfigPath := filepath.Join(defaultAirshipConfigDir, config.AirshipConfig)
	flags.StringVar(
		&a.AirshipConfigPath,
		"airshipconf",
		"",
		`Path to file for airshipctl configuration. (default "`+defaultAirshipConfigPath+`")`)

	defaultKubeConfigPath := filepath.Join(defaultAirshipConfigDir, config.AirshipKubeConfig)
	flags.StringVar(
		&a.KubeConfigPath,
		clientcmd.RecommendedConfigPathFlag,
		"",
		`Path to kubeconfig associated with airshipctl configuration. (default "`+defaultKubeConfigPath+`")`)

	a.Create = false
}

// InitConfig - Initializes and loads Config it exists.
func (a *AirshipCTLSettings) InitConfig() {
	a.Config = config.NewConfig()

	a.InitAirshipConfigPath()
	a.InitKubeConfigPath()
	InitPluginPath()

	err := a.Config.LoadConfig(a.AirshipConfigPath, a.KubeConfigPath, a.Create)
	if err != nil {
		// Should stop airshipctl
		log.Fatal("Failed to load or initialize config: ", err)
	}
}

// InitAirshipConfigPath - Initializes AirshipConfigPath variable for Config object
func (a *AirshipCTLSettings) InitAirshipConfigPath() {
	// The airshipConfigPath may already have been received as a command line argument
	if a.AirshipConfigPath != "" {
		return
	}

	// Otherwise, we can check if we got the path via ENVIRONMENT variable
	a.AirshipConfigPath = os.Getenv(config.AirshipConfigEnv)
	if a.AirshipConfigPath != "" {
		return
	}

	// Otherwise, we'll try putting it in the home directory
	homeDir := userHomeDir()
	a.AirshipConfigPath = filepath.Join(homeDir, config.AirshipConfigDir, config.AirshipConfig)
}

// InitKubeConfigPath - Initializes KubeConfigPath variable for Config object
func (a *AirshipCTLSettings) InitKubeConfigPath() {
	// NOTE(howell): This function will set the kubeConfigPath to the
	// default location under the airship directory unless the user
	// *explicitly* specifies a different location, either by setting the
	// ENVIRONMENT variable or by passing a command line argument.
	// This avoids messing up the user's kubeconfig if they didn't
	// explicitly want airshipctl to use it.

	// The kubeConfigPath may already have been received as a command line argument
	if a.KubeConfigPath != "" {
		return
	}

	// Otherwise, we can check if we got the path via ENVIRONMENT variable
	a.KubeConfigPath = os.Getenv(config.AirshipKubeConfigEnv)
	if a.KubeConfigPath != "" {
		return
	}

	// Otherwise, we'll try putting it in the home directory
	homeDir := userHomeDir()
	a.KubeConfigPath = filepath.Join(homeDir, config.AirshipConfigDir, config.AirshipKubeConfig)
}

// InitPluginPath - Sets the location to look for kustomize plugins (including airshipctl itself).
func InitPluginPath() {
	pluginPathLock.Lock()
	defer pluginPathLock.Unlock()

	// Check if we got the path via ENVIRONMENT variable
	pluginPath = os.Getenv(config.AirshipPluginPathEnv)
	if pluginPath != "" {
		return
	}

	// Otherwise, we'll try putting it in the home directory
	homeDir := userHomeDir()
	pluginPath = filepath.Join(homeDir, config.AirshipConfigDir, config.AirshipPluginPath)
}

// PluginPath returns the kustomize plugin path
func PluginPath() string {
	pluginPathLock.Lock()
	defer pluginPathLock.Unlock()
	return pluginPath
}

// userHomeDir is a utility function that wraps os.UserHomeDir and returns no
// errors. If the user has no home directory, the returned value will be the
// empty string
func userHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}
	return homeDir
}
