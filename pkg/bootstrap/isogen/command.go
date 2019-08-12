package isogen

import (
	"context"
	"fmt"
	"io"

	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/util"
)

// GenerateBootstrapIso will generate data for cloud init and start ISO builder container
func GenerateBootstrapIso(settings *Settings, args []string, out io.Writer) error {
	if settings.IsogenConfigFile == "" {
		fmt.Fprintln(out, "Reading config file location from global settings is not supported")
		return errors.ErrNotImplemented{}
	}

	ctx := context.Background()
	cfg := &Config{}

	if err := util.ReadYAMLFile(settings.IsogenConfigFile, &cfg); err != nil {
		return err
	}

	fmt.Fprintln(out, "Creating ISO builder container")
	builder, err := container.NewContainer(
		&ctx, cfg.Container.ContainerRuntime,
		cfg.Container.Image)
	if err != nil {
		return err
	}

	return generateBootstrapIso(builder, cfg, out, settings.Debug)
}

func generateBootstrapIso(builder container.Container, cfg *Config, out io.Writer, debug bool) error {
	vols := []string{cfg.Container.Volume}
	fmt.Fprintf(out, "Running default container command. Mounted dir: %s\n", vols)
	if err := builder.RunCommand(
		[]string{},
		nil,
		vols,
		[]string{},
		debug,
	); err != nil {
		return err
	}

	fmt.Fprintln(out, "ISO successfully built.")
	if debug {
		fmt.Fprintf(
			out,
			"Debug flag is set. Container %s stopped but not deleted.\n",
			builder.GetId(),
		)
		return nil
	}
	fmt.Fprintln(out, "Removing container.")
	return builder.RmContainer()
}
