package isogen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"opendev.org/airship/airshipctl/pkg/bootstrap/cloudinit"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/util"

	"sigs.k8s.io/kustomize/v3/pkg/fs"
)

const (
	builderConfigFileName = "builder-conf.yaml"
)

// GenerateBootstrapIso will generate data for cloud init and start ISO builder container
func GenerateBootstrapIso(settings *Settings, args []string) error {
	if settings.IsogenConfigFile == "" {
		log.Print("Reading config file location from global settings is not supported")
		return errors.ErrNotImplemented{}
	}

	ctx := context.Background()
	cfg := &Config{}

	if err := util.ReadYAMLFile(settings.IsogenConfigFile, &cfg); err != nil {
		return err
	}

	if err := verifyInputs(cfg, args); err != nil {
		return err
	}

	docBundle, err := document.NewBundle(fs.MakeRealFS(), args[0], "")
	if err != nil {
		return err
	}

	log.Print("Creating ISO builder container")
	builder, err := container.NewContainer(
		&ctx, cfg.Container.ContainerRuntime,
		cfg.Container.Image)
	if err != nil {
		return err
	}

	err = generateBootstrapIso(docBundle, builder, cfg, settings.Debug)
	if err != nil {
		return err
	}
	log.Print("Checking artifacts")
	return verifyArtifacts(cfg)
}

func verifyInputs(cfg *Config, args []string) error {
	if len(args) == 0 {
		log.Print("Specify path to document model. Config param from global settings is not supported")
		return errors.ErrNotImplemented{}
	}

	if cfg.Container.Volume == "" {
		log.Print("Specify volume bind for ISO builder container")
		return errors.ErrWrongConfig{}
	}

	if (cfg.Builder.UserDataFileName == "") || (cfg.Builder.NetworkConfigFileName == "") {
		log.Print("UserDataFileName or NetworkConfigFileName are not specified in ISO builder config")
		return errors.ErrWrongConfig{}
	}

	vols := strings.Split(cfg.Container.Volume, ":")
	switch {
	case len(vols) == 1:
		cfg.Container.Volume = fmt.Sprintf("%s:%s", vols[0], vols[0])
	case len(vols) > 2:
		log.Print("Bad container volume format. Use hostPath:contPath")
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

func verifyArtifacts(cfg *Config) error {
	hostVol := strings.Split(cfg.Container.Volume, ":")[0]
	metadataPath := filepath.Join(hostVol, cfg.Builder.OutputMetadataFileName)
	_, err := os.Stat(metadataPath)
	return err
}

func generateBootstrapIso(
	docBubdle document.Bundle,
	builder container.Container,
	cfg *Config,
	debug bool,
) error {
	cntVol := strings.Split(cfg.Container.Volume, ":")[1]
	log.Print("Creating cloud-init for ephemeral K8s")
	userData, netConf, err := cloudinit.GetCloudData(docBubdle, EphemeralClusterAnnotation)
	if err != nil {
		return err
	}

	var fls map[string][]byte
	fls, err = getContainerCfg(cfg, userData, netConf)
	if err != nil {
		return err
	}

	if err = util.WriteFiles(fls, 0600); err != nil {
		return err
	}

	vols := []string{cfg.Container.Volume}
	builderCfgLocation := filepath.Join(cntVol, builderConfigFileName)
	log.Printf("Running default container command. Mounted dir: %s", vols)
	if err := builder.RunCommand(
		[]string{},
		nil,
		vols,
		[]string{fmt.Sprintf("BUILDER_CONFIG=%s", builderCfgLocation)},
		debug,
	); err != nil {
		return err
	}

	log.Print("ISO successfully built.")
	if !debug {
		log.Print("Removing container.")
		return builder.RmContainer()
	}

	log.Debugf("Debug flag is set. Container %s stopped but not deleted.", builder.GetId())
	return nil
}
