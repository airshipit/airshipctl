package environment

import (
	"github.com/spf13/cobra"
)

type PluginSettings interface {
	Init() error
}

// AirshipCTLSettings is a container for all of the settings needed by airshipctl
type AirshipCTLSettings struct {
	// Debug is used for verbose output
	Debug bool

	PluginSettings map[string]PluginSettings
}

// InitFlags adds the default settings flags to cmd
func (a *AirshipCTLSettings) InitFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.BoolVar(&a.Debug, "debug", false, "enable verbose output")
}

func (a *AirshipCTLSettings) Register(pluginName string, pluginSettings PluginSettings) {
	if a.PluginSettings == nil {
		a.PluginSettings = make(map[string]PluginSettings, 0)
	}
	a.PluginSettings[pluginName] = pluginSettings
}

func (a *AirshipCTLSettings) Init() error {
	for _, pluginSettings := range a.PluginSettings {
		if err := pluginSettings.Init(); err != nil {
			return err
		}
	}
	return nil
}
