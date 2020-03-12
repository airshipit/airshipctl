package isogen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"opendev.org/airship/airshipctl/pkg/bootstrap/cloudinit"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/util"
)

const (
	builderConfigFileName = "builder-conf.yaml"
)

// GenerateBootstrapIso will generate data for cloud init and start ISO builder container
func GenerateBootstrapIso(settings *environment.AirshipCTLSettings) error {
	ctx := context.Background()

	globalConf := settings.Config()
	if err := globalConf.EnsureComplete(); err != nil {
		return err
	}

	cfg, err := globalConf.CurrentContextBootstrapInfo()
	if err != nil {
		return err
	}

	if err = verifyInputs(cfg); err != nil {
		return err
	}

	// TODO (dukov) replace with the appropriate function once it's available
	// in document module
	root, err := globalConf.CurrentContextEntryPoint(config.Ephemeral, "")
	if err != nil {
		return err
	}
	docBundle, err := document.NewBundleByPath(root)
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

func verifyInputs(cfg *config.Bootstrap) error {
	if cfg.Container.Volume == "" {
		return config.ErrMissingConfig{
			What: "Must specify volume bind for ISO builder container",
		}
	}

	if (cfg.Builder.UserDataFileName == "") || (cfg.Builder.NetworkConfigFileName == "") {
		return config.ErrMissingConfig{
			What: "UserDataFileName or NetworkConfigFileName are not specified in ISO builder config",
		}
	}

	vols := strings.Split(cfg.Container.Volume, ":")
	switch {
	case len(vols) == 1:
		cfg.Container.Volume = fmt.Sprintf("%s:%s", vols[0], vols[0])
	case len(vols) > 2:
		return config.ErrInvalidConfig{
			What: "Bad container volume format. Use hostPath:contPath",
		}
	}
	return nil
}

func getContainerCfg(cfg *config.Bootstrap, userData []byte, netConf []byte) map[string][]byte {
	hostVol := strings.Split(cfg.Container.Volume, ":")[0]

	fls := make(map[string][]byte)
	fls[filepath.Join(hostVol, cfg.Builder.UserDataFileName)] = userData
	fls[filepath.Join(hostVol, cfg.Builder.NetworkConfigFileName)] = netConf
	// TODO (dukov) Get rid of this ugly conversion byte -> string -> byte
	builderData := []byte(cfg.String())
	fls[filepath.Join(hostVol, builderConfigFileName)] = builderData
	return fls
}

func verifyArtifacts(cfg *config.Bootstrap) error {
	hostVol := strings.Split(cfg.Container.Volume, ":")[0]
	metadataPath := filepath.Join(hostVol, cfg.Builder.OutputMetadataFileName)
	_, err := os.Stat(metadataPath)
	return err
}

func generateBootstrapIso(
	docBundle document.Bundle,
	builder container.Container,
	cfg *config.Bootstrap,
	debug bool,
) error {
	cntVol := strings.Split(cfg.Container.Volume, ":")[1]
	log.Print("Creating cloud-init for ephemeral K8s")
	userData, netConf, err := cloudinit.GetCloudData(docBundle)
	if err != nil {
		return err
	}

	fls := getContainerCfg(cfg, userData, netConf)
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
		[]string{
			fmt.Sprintf("BUILDER_CONFIG=%s", builderCfgLocation),
			fmt.Sprintf("http_proxy=%s", os.Getenv("http_proxy")),
			fmt.Sprintf("https_proxy=%s", os.Getenv("https_proxy")),
			fmt.Sprintf("HTTP_PROXY=%s", os.Getenv("HTTP_PROXY")),
			fmt.Sprintf("HTTPS_PROXY=%s", os.Getenv("HTTPS_PROXY")),
			fmt.Sprintf("NO_PROXY=%s", os.Getenv("NO_PROXY")),
		},
		debug,
	); err != nil {
		return err
	}

	log.Print("ISO successfully built.")
	if !debug {
		log.Print("Removing container.")
		return builder.RmContainer()
	}

	log.Debugf("Debug flag is set. Container %s stopped but not deleted.", builder.GetID())
	return nil
}
