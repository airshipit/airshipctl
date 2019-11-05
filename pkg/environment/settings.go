package environment

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"k8s.io/client-go/tools/clientcmd"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/log"
)

// AirshipCTLSettings is a container for all of the settings needed by airshipctl
type AirshipCTLSettings struct {
	// Debug is used for verbose output
	Debug             bool
	airshipConfigPath string
	kubeConfigPath    string
	config            *config.Config
}

// InitFlags adds the default settings flags to cmd
func (a *AirshipCTLSettings) InitFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.BoolVar(&a.Debug, "debug", false, "enable verbose output")

	flags.StringVar(&a.airshipConfigPath, config.FlagConfigFilePath,
		filepath.Join(HomePlaceholder, config.AirshipConfigDir, config.AirshipConfig),
		"Path to file for airshipctl configuration.")

	flags.StringVar(&a.kubeConfigPath, clientcmd.RecommendedConfigPathFlag,
		filepath.Join(HomePlaceholder, config.AirshipConfigDir, config.AirshipKubeConfig),
		"Path to kubeconfig associated with airshipctl configuration.")

}

func (a *AirshipCTLSettings) Config() *config.Config {
	return a.config
}
func (a *AirshipCTLSettings) SetConfig(conf *config.Config) {
	a.config = conf
}

func (a *AirshipCTLSettings) AirshipConfigPath() string {
	return a.airshipConfigPath
}
func (a *AirshipCTLSettings) SetAirshipConfigPath(acp string) {
	a.airshipConfigPath = acp
}
func (a *AirshipCTLSettings) KubeConfigPath() string {
	return a.kubeConfigPath
}
func (a *AirshipCTLSettings) SetKubeConfigPath(kcp string) {
	a.kubeConfigPath = kcp
}

// InitConfig - Initializes and loads Config it exists.
func (a *AirshipCTLSettings) InitConfig() {

	// Raw - Empty Config object
	a.SetConfig(config.NewConfig())

	a.setAirshipConfigPath()
	//Pass the airshipConfigPath and kubeConfig object
	err := a.Config().LoadConfig(a.AirshipConfigPath(), a.setKubePathOptions())
	if err != nil {
		// Should stop airshipctl
		log.Fatal(err)
	}

}

func (a *AirshipCTLSettings) setAirshipConfigPath() {
	// (1) If the airshipConfigPath was received as an argument its aleady set
	if a.airshipConfigPath == "" {
		// (2) If not , we can check if we got the Path via ENVIRONMNT variable,
		// set the appropriate fields
		a.setAirshipConfigPathFromEnv()
	}
	// (3) Check if the a.airshipConfigPath is empty still at this point , use the defaults
	acp, home := a.replaceHomePlaceholder(a.airshipConfigPath)
	a.airshipConfigPath = acp
	if a.airshipConfigPath == "" {
		a.airshipConfigPath = filepath.Join(home, config.AirshipConfigDir, config.AirshipConfig)
	}
}

// setAirshipConfigPathFromEnv Get AIRSHIP CONFIG from an environment variable if set
func (a *AirshipCTLSettings) setAirshipConfigPathFromEnv() {
	// Check if AIRSHIPCONF env variable was set
	// I have the path and name for the airship config file
	a.airshipConfigPath = os.Getenv(config.AirshipConfigEnv)
}

func (a *AirshipCTLSettings) setKubePathOptions() *clientcmd.PathOptions {
	// USe default expectations for Kubeconfig
	kubePathOptions := clientcmd.NewDefaultPathOptions()
	// No need to check the Environment , since we are relying on the kubeconfig defaults
	// If we did not get an explicit kubeconfig definition on airshipctl
	// as far as airshipctkl is concerned will use the default expectations for the kubeconfig
	// file location . This avoids messing up someones kubeconfig if they didnt explicitly want
	// airshipctl to use it.
	kcp, home := a.replaceHomePlaceholder(a.kubeConfigPath)
	a.kubeConfigPath = kcp
	if a.kubeConfigPath == "" {
		a.kubeConfigPath = filepath.Join(home, config.AirshipConfigDir, config.AirshipKubeConfig)
	}
	//  We will always rely on tha airshipctl cli args or default for where to find kubeconfig
	kubePathOptions.GlobalFile = a.kubeConfigPath
	kubePathOptions.GlobalFileSubpath = a.kubeConfigPath

	return kubePathOptions

}
func (a *AirshipCTLSettings) replaceHomePlaceholder(configPath string) (string, string) {
	home, err := os.UserHomeDir()
	if err != nil {
		// Use defaults under current directory
		home = ""
	}
	if configPath == "" {
		return configPath, home
	}

	return strings.Replace(configPath, HomePlaceholder, home, 1), home
}
