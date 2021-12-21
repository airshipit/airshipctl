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

package executors

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	airerrors "opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/log"
	phaseerrors "opendev.org/airship/airshipctl/pkg/phase/errors"
	"opendev.org/airship/airshipctl/pkg/phase/executors/errors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

const clusterAPIOverrides = "/workdir/.cluster-api/overrides"

var _ ifc.Executor = &ClusterctlExecutor{}

// ClusterctlExecutor phase executor
type ClusterctlExecutor struct {
	clusterName string
	targetPath  string

	clusterMap clustermap.ClusterMap
	options    *airshipv1.Clusterctl
	kubecfg    kubeconfig.Interface
	execObj    *airshipv1.GenericContainer
	clientFunc container.ClientV1Alpha1FactoryFunc
	cctlOpts   *airshipv1.ClusterctlOptions
}

var typeMap = map[string]string{
	airshipv1.BootstrapProviderType:      "bootstrap",
	airshipv1.ControlPlaneProviderType:   "control-plane",
	airshipv1.InfrastructureProviderType: "infrastructure",
	airshipv1.CoreProviderType:           "core",
}

// NewClusterctlExecutor creates instance of 'clusterctl' phase executor
func NewClusterctlExecutor(cfg ifc.ExecutorConfig) (ifc.Executor, error) {
	options := airshipv1.DefaultClusterctl()
	if err := cfg.ExecutorDocument.ToAPIObject(options, airshipv1.Scheme); err != nil {
		return nil, err
	}
	cctlOpts := &airshipv1.ClusterctlOptions{
		Components: map[string]string{},
	}
	if err := initRepoData(options, cctlOpts, cfg.TargetPath); err != nil {
		return nil, err
	}

	doc, err := cfg.PhaseConfigBundle.SelectOne(document.NewClusterctlContainerExecutorSelector())
	if err != nil {
		return nil, err
	}

	apiObj := airshipv1.DefaultGenericContainer()
	err = doc.ToAPIObject(apiObj, airshipv1.Scheme)
	if err != nil {
		return nil, err
	}

	clientFunc := container.NewClientV1Alpha1
	if cfg.ContainerFunc != nil {
		clientFunc = cfg.ContainerFunc
	}

	return &ClusterctlExecutor{
		clusterName: cfg.ClusterName,
		options:     options,
		cctlOpts:    cctlOpts,
		kubecfg:     cfg.KubeConfig,
		clusterMap:  cfg.ClusterMap,
		targetPath:  cfg.TargetPath,
		execObj:     apiObj,
		clientFunc:  clientFunc,
	}, nil
}

func initRepoData(c *airshipv1.Clusterctl, o *airshipv1.ClusterctlOptions, targetPath string) error {
	for _, prv := range c.Providers {
		rURL, err := url.Parse(prv.URL)
		if err != nil {
			return err
		}
		if rURL.Scheme != "" || filepath.IsAbs(prv.URL) {
			continue
		}

		componentDir := filepath.Join(clusterAPIOverrides,
			fmt.Sprintf("%s-%s", typeMap[prv.Type], prv.Name), filepath.Base(prv.URL))
		if prv.Type == airshipv1.CoreProviderType {
			componentDir = filepath.Join(clusterAPIOverrides, prv.Name, filepath.Base(prv.URL))
		}

		kustomizePath := filepath.Join(targetPath, prv.URL)
		log.Debugf("Building cluster-api provider component documents from kustomize path at '%s'", kustomizePath)
		bundle, err := document.NewBundleByPath(kustomizePath)
		if err != nil {
			return err
		}
		doc, err := bundle.SelectOne(document.NewClusterctlMetadataSelector())
		if err != nil {
			return err
		}
		metadata, err := doc.AsYAML()
		if err != nil {
			return err
		}

		o.Components[filepath.Join(componentDir, "metadata.yaml")] = string(metadata)

		filteredBundle, err := bundle.SelectBundle(document.NewDeployToK8sSelector())
		if err != nil {
			return err
		}

		buffer := &bytes.Buffer{}
		if err = filteredBundle.Write(buffer); err != nil {
			return err
		}
		prv.URL = filepath.Join(componentDir, fmt.Sprintf("%s-components.yaml", typeMap[prv.Type]))
		o.Components[prv.URL] = string(buffer.Bytes())
	}
	return nil
}

// Run clusterctl init as a phase runner
func (c *ClusterctlExecutor) Run(opts ifc.RunOptions) error {
	if log.DebugEnabled() {
		c.cctlOpts.CmdOptions = append(c.cctlOpts.CmdOptions, "-v5")
	}

	cctlConfig := map[string]interface{}{
		"providers": c.options.Providers,
		"images":    c.options.ImageMetas,
	}
	for k, v := range c.options.AdditionalComponentVariables {
		cctlConfig[k] = v
	}

	var err error
	c.cctlOpts.Config, err = yaml.Marshal(cctlConfig)
	if err != nil {
		return err
	}

	switch c.options.Action {
	case airshipv1.Init:
		return c.init()
	case airshipv1.Move:
		return c.move(opts.DryRun)
	default:
		return errors.ErrUnknownExecutorAction{Action: string(c.options.Action), ExecutorName: "clusterctl"}
	}
}

func (c *ClusterctlExecutor) run() error {
	opts, err := yaml.Marshal(c.cctlOpts)
	if err != nil {
		return err
	}
	c.execObj.Config = string(opts)
	return c.clientFunc("", &bytes.Buffer{}, os.Stdout, c.execObj, c.targetPath).Run()
}

func (c *ClusterctlExecutor) getKubeconfig() (string, string, func(), error) {
	kubeConfigFile, cleanup, err := c.kubecfg.GetFile()
	if err != nil {
		return "", "", nil, err
	}

	context, err := c.clusterMap.ClusterKubeconfigContext(c.clusterName)
	if err != nil {
		cleanup()
		return "", "", nil, err
	}

	c.execObj.Spec.StorageMounts = append(c.execObj.Spec.StorageMounts, airshipv1.StorageMount{
		MountType:     "bind",
		Src:           kubeConfigFile,
		DstPath:       kubeConfigFile,
		ReadWriteMode: false,
	})
	return kubeConfigFile, context, cleanup, nil
}

func (c *ClusterctlExecutor) init() error {
	log.Print("starting clusterctl init executor")

	kubecfg, context, cleanup, err := c.getKubeconfig()
	if err != nil {
		return err
	}
	defer cleanup()

	c.cctlOpts.CmdOptions = append(c.cctlOpts.CmdOptions,
		"init",
		"--kubeconfig", kubecfg,
		"--kubeconfig-context", context,
	)

	initMap := map[string]string{
		typeMap[airshipv1.BootstrapProviderType]:      c.options.InitOptions.BootstrapProviders,
		typeMap[airshipv1.ControlPlaneProviderType]:   c.options.InitOptions.ControlPlaneProviders,
		typeMap[airshipv1.InfrastructureProviderType]: c.options.InitOptions.InfrastructureProviders,
		typeMap[airshipv1.CoreProviderType]:           c.options.InitOptions.CoreProvider,
	}
	for k, v := range initMap {
		if v != "" {
			c.cctlOpts.CmdOptions = append(c.cctlOpts.CmdOptions, fmt.Sprintf("--%s=%s", k, v))
		}
	}

	if err = c.run(); err != nil {
		return err
	}

	log.Print("clusterctl init completed successfully")
	return nil
}

func (c *ClusterctlExecutor) move(dryRun bool) error {
	log.Print("starting clusterctl move executor")

	kubecfg, context, cleanup, err := c.getKubeconfig()
	if err != nil {
		return err
	}
	defer cleanup()

	fromCluster, err := c.clusterMap.ParentCluster(c.clusterName)
	if err != nil {
		return err
	}
	fromContext, err := c.clusterMap.ClusterKubeconfigContext(fromCluster)
	if err != nil {
		return err
	}

	c.cctlOpts.CmdOptions = append(
		c.cctlOpts.CmdOptions,
		"move",
		"--kubeconfig", kubecfg,
		"--kubeconfig-context", fromContext,
		"--to-kubeconfig", kubecfg,
		"--to-kubeconfig-context", context,
		"--namespace", c.options.MoveOptions.Namespace,
	)

	if dryRun {
		c.cctlOpts.CmdOptions = append(
			c.cctlOpts.CmdOptions,
			"--dry-run",
		)
	}

	if err = c.run(); err != nil {
		return err
	}

	log.Print("clusterctl move completed successfully")
	return nil
}

// Validate executor configuration and documents
func (c *ClusterctlExecutor) Validate() error {
	switch c.options.Action {
	case "":
		return phaseerrors.ErrInvalidPhase{Reason: "ClusterctlExecutor.Action is empty"}
	case airshipv1.Init:
		if c.options.InitOptions.CoreProvider == "" {
			log.Printf("ClusterctlExecutor.InitOptions.CoreProvider is empty")
		}
	case airshipv1.Move:
	default:
		return errors.ErrUnknownExecutorAction{Action: string(c.options.Action)}
	}
	// TODO: need to find if any other validation needs to be added
	return nil
}

// Render executor documents
func (c *ClusterctlExecutor) Render(w io.Writer, ro ifc.RenderOptions) error {
	dataAll := &bytes.Buffer{}
	for path, data := range c.cctlOpts.Components {
		if strings.Contains(path, "components.yaml") {
			dataAll.Write(append([]byte(data), []byte("\n---\n")...))
		}
	}

	bundle, err := document.NewBundleFromBytes(dataAll.Bytes())
	if err != nil {
		return err
	}
	filtered, err := bundle.SelectBundle(ro.FilterSelector)
	if err != nil {
		return err
	}
	return filtered.Write(w)
}

// Status returns the status of the given phase
func (c *ClusterctlExecutor) Status() (ifc.ExecutorStatus, error) {
	return ifc.ExecutorStatus{}, airerrors.ErrNotImplemented{What: Clusterctl}
}
