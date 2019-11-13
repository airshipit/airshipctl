package isogen

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"opendev.org/airship/airshipctl/pkg/bootstrap/cloudinit"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/util"

	"sigs.k8s.io/kustomize/v3/pkg/fs"
)

const (
	builderConfigFileName = "builder-conf.yaml"
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

	if err := verifyInputs(cfg, args, out); err != nil {
		return err
	}

	docBundle, err := document.NewBundle(fs.MakeRealFS(), args[0], "")
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "Creating ISO builder container")
	builder, err := container.NewContainer(
		&ctx, cfg.Container.ContainerRuntime,
		cfg.Container.Image)
	if err != nil {
		return err
	}

	return generateBootstrapIso(docBundle, builder, cfg, out, settings.Debug)
}

func verifyInputs(cfg *Config, args []string, out io.Writer) error {
	if len(args) == 0 {
		fmt.Fprintln(out, "Specify path to document model. Config param from global settings is not supported")
		return errors.ErrNotImplemented{}
	}

	if cfg.Container.Volume == "" {
		fmt.Fprintln(out, "Specify volume bind for ISO builder container")
		return errors.ErrWrongConfig{}
	}

	if (cfg.Builder.UserDataFileName == "") || (cfg.Builder.NetworkConfigFileName == "") {
		fmt.Fprintln(out, "UserDataFileName or NetworkConfigFileName are not specified in ISO builder config")
		return errors.ErrWrongConfig{}
	}

	vols := strings.Split(cfg.Container.Volume, ":")
	switch {
	case len(vols) == 1:
		cfg.Container.Volume = fmt.Sprintf("%s:%s", vols[0], vols[0])
	case len(vols) > 2:
		fmt.Fprintln(out, "Bad container volume format. Use hostPath:contPath")
		return errors.ErrWrongConfig{}
	}
	return nil
}

func getContainerCfg(cfg *Config, userData []byte, netConf []byte) (map[string][]byte, error) {
	hostVol := strings.Split(cfg.Container.Volume, ":")[0]

	fls := make(map[string][]byte)
	fls[filepath.Join(hostVol, cfg.Builder.UserDataFileName)] = userData
	fls[filepath.Join(hostVol, cfg.Builder.NetworkConfigFileName)] = netConf
	builderData, err := cfg.ToYAML()
	if err != nil {
		return nil, err
	}
	fls[filepath.Join(hostVol, builderConfigFileName)] = builderData
	return fls, nil
}

func generateBootstrapIso(
	docBubdle document.Bundle,
	builder container.Container,
	cfg *Config,
	out io.Writer,
	debug bool,
) error {
	cntVol := strings.Split(cfg.Container.Volume, ":")[1]
	fmt.Fprintln(out, "Creating cloud-init for ephemeral K8s")
	userData, netConf, err := cloudinit.GetCloudData(docBubdle, EphemeralClusterAnnotation)
	if err != nil {
		return err
	}

	var fls map[string][]byte
	fls, err = getContainerCfg(cfg, userData, netConf)
	if err = util.WriteFiles(fls, 0600); err != nil {
		return err
	}

	vols := []string{cfg.Container.Volume}
	builderCfgLocation := filepath.Join(cntVol, builderConfigFileName)
	fmt.Fprintf(out, "Running default container command. Mounted dir: %s\n", vols)
	if err := builder.RunCommand(
		[]string{},
		nil,
		vols,
		[]string{fmt.Sprintf("BUILDER_CONFIG=%s", builderCfgLocation)},
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
