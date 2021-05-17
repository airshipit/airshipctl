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
	goerrors "errors"
	"io"
	"os"
	"path/filepath"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	commonerrors "opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/phase/errors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

var _ ifc.Executor = &ContainerExecutor{}

// ContainerExecutor contains resources for generic container executor
type ContainerExecutor struct {
	ResultsDir    string
	MountBasePath string

	Container        *v1alpha1.GenericContainer
	ClientFunc       container.ClientV1Alpha1FactoryFunc
	ExecutorBundle   document.Bundle
	ExecutorDocument document.Document
	Options          ifc.ExecutorConfig
}

// NewContainerExecutor creates instance of phase executor
func NewContainerExecutor(cfg ifc.ExecutorConfig) (ifc.Executor, error) {
	bundle, err := cfg.BundleFactory()
	// ErrDocumentEntrypointNotDefined error should not cause Container executor to fail, so filter it
	if err != nil && goerrors.As(err, &errors.ErrDocumentEntrypointNotDefined{}) {
		// if docEntryPoint isn't defined initialize empty bundle instead to safely use it without nil checks
		bundle, err = document.NewBundleFromBytes([]byte{})
	}
	if err != nil {
		return nil, err
	}

	apiObj := v1alpha1.DefaultGenericContainer()
	err = cfg.ExecutorDocument.ToAPIObject(apiObj, v1alpha1.Scheme)
	if err != nil {
		return nil, err
	}

	var resultsDir string
	if apiObj.Spec.SinkOutputDir != "" {
		resultsDir = filepath.Join(cfg.SinkBasePath, apiObj.Spec.SinkOutputDir)
	}

	return &ContainerExecutor{
		ResultsDir:       resultsDir,
		MountBasePath:    cfg.TargetPath,
		ExecutorBundle:   bundle,
		ExecutorDocument: cfg.ExecutorDocument,
		// TODO extend tests with proper client, make it interface
		ClientFunc: container.NewClientV1Alpha1,
		Container:  apiObj,
		Options:    cfg,
	}, nil
}

// Run generic container as a phase runner
func (c *ContainerExecutor) Run(evtCh chan events.Event, opts ifc.RunOptions) {
	defer close(evtCh)

	evtCh <- events.NewEvent().WithGenericContainerEvent(events.GenericContainerEvent{
		Operation: events.GenericContainerStart,
		Message:   "starting generic container",
	})

	if c.Options.ClusterName != "" {
		cleanup, err := c.SetKubeConfig()
		if err != nil {
			handleError(evtCh, err)
			return
		}
		defer cleanup()
	}

	input, err := bundleReader(c.ExecutorBundle)
	if err != nil {
		handleError(evtCh, err)
		return
	}

	// TODO this logic is redundant in executor package, move it to pkg/container
	var output io.Writer
	if c.ResultsDir == "" {
		// set output only if the output if resulting directory is not defined
		output = os.Stdout
	}
	if err = c.setConfig(); err != nil {
		handleError(evtCh, err)
		return
	}

	// TODO check the executor type  when dryrun is set
	if opts.DryRun {
		evtCh <- events.NewEvent().WithGenericContainerEvent(events.GenericContainerEvent{
			Operation: events.GenericContainerStop,
			Message:   "DryRun execution finished",
		})
		return
	}

	err = c.ClientFunc(c.ResultsDir, input, output, c.Container, c.MountBasePath).Run()
	if err != nil {
		handleError(evtCh, err)
		return
	}

	evtCh <- events.NewEvent().WithGenericContainerEvent(events.GenericContainerEvent{
		Operation: events.GenericContainerStop,
		Message:   "execution of the generic container finished",
	})
}

// SetKubeConfig adds env variable and mounts kubeconfig to container
func (c *ContainerExecutor) SetKubeConfig() (kubeconfig.Cleanup, error) {
	context, err := c.Options.ClusterMap.ClusterKubeconfigContext(c.Options.ClusterName)
	if err != nil {
		return nil, err
	}
	kubeConfigSrc, cleanup, err := c.Options.KubeConfig.GetFile()
	if err != nil {
		return nil, err
	}
	c.Container.Spec.StorageMounts = append(c.Container.Spec.StorageMounts, v1alpha1.StorageMount{
		MountType: "bind",
		Src:       kubeConfigSrc,
		DstPath:   v1alpha1.KubeConfigPath,
	})
	envs := []string{v1alpha1.KubeConfigEnv, v1alpha1.KubeConfigEnvKeyContext + "=" + context}
	c.Container.Spec.EnvVars = append(c.Container.Spec.EnvVars, envs...)

	return cleanup, nil
}

// bundleReader sets input for function
func bundleReader(bundle document.Bundle) (io.Reader, error) {
	buf := &bytes.Buffer{}
	return buf, bundle.Write(buf)
}

// Validate executor configuration and documents
func (c *ContainerExecutor) Validate() error {
	log.Printf("Method Validate() for container isn't implemented")
	return nil
}

// Render executor documents
func (c *ContainerExecutor) Render(w io.Writer, o ifc.RenderOptions) error {
	bundle, err := c.ExecutorBundle.SelectBundle(o.FilterSelector)
	if err != nil {
		return err
	}
	return bundle.Write(w)
}

func (c *ContainerExecutor) setConfig() error {
	if c.Container.ConfigRef != nil {
		log.Debugf("Config reference is specified, looking for the object in config ref: '%v'", c.Container.ConfigRef)
		gvk := c.Container.ConfigRef.GroupVersionKind()
		selector := document.NewSelector().
			ByName(c.Container.ConfigRef.Name).
			ByNamespace(c.Container.ConfigRef.Namespace).
			ByGvk(gvk.Group, gvk.Version, gvk.Kind)
		doc, err := c.Options.PhaseConfigBundle.SelectOne(selector)
		if err != nil {
			return err
		}
		config, err := doc.AsYAML()
		if err != nil {
			return err
		}
		c.Container.Config = string(config)
		return nil
	}
	return nil
}

// Status returns the status of the given phase
func (c *ContainerExecutor) Status() (ifc.ExecutorStatus, error) {
	return ifc.ExecutorStatus{}, commonerrors.ErrNotImplemented{What: GenericContainer}
}
