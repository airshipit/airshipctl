package isogen

import (
	"errors"
	"fmt"
	"io"

	"opendev.org/airship/airshipctl/pkg/util"
)

// ErrNotImplemented returned for not implemented features
var ErrNotImplemented = errors.New("Error. Not implemented")

// GenerateBootstrapIso will generate data for cloud init and start ISO builder container
func GenerateBootstrapIso(settings *Settings, args []string, out io.Writer) error {
	if settings.IsogenConfigFile == "" {
		fmt.Fprintln(out, "Reading config file location from global settings is not supported")
		return ErrNotImplemented
	}

	cfg := Config{}

	if err := util.ReadYAMLFile(settings.IsogenConfigFile, &cfg); err != nil {
		return err
	}
	fmt.Println("Under construction")
	return nil
}
