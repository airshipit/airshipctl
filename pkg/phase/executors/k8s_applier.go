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
	"io"
	"os"

	"sigs.k8s.io/kustomize/kyaml/yaml"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	airerrors "opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/phase/errors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

var _ ifc.Executor = &KubeApplierExecutor{}

// KubeApplierExecutor applies resources to kubernetes
type KubeApplierExecutor struct {
	ExecutorBundle   document.Bundle
	ExecutorDocument document.Document
	BundleName       string
	targetPath       string

	apiObject   *airshipv1.KubernetesApply
	clusterMap  clustermap.ClusterMap
	clusterName string
	kubeconfig  kubeconfig.Interface
	clientFunc  container.ClientV1Alpha1FactoryFunc
	execObj     *airshipv1.GenericContainer
}

// NewKubeApplierExecutor returns instance of executor
func NewKubeApplierExecutor(cfg ifc.ExecutorConfig) (ifc.Executor, error) {
	apiObj := &airshipv1.KubernetesApply{}
	err := cfg.ExecutorDocument.ToAPIObject(apiObj, airshipv1.Scheme)
	if err != nil {
		return nil, err
	}
	bundle, err := cfg.BundleFactory()
	if err != nil {
		return nil, err
	}

	doc, err := cfg.PhaseConfigBundle.SelectOne(document.NewApplierContainerExecutorSelector())
	if err != nil {
		return nil, err
	}

	cObj := airshipv1.DefaultGenericContainer()
	err = doc.ToAPIObject(cObj, airshipv1.Scheme)
	if err != nil {
		return nil, err
	}

	clientFunc := container.NewClientV1Alpha1
	if cfg.ContainerFunc != nil {
		clientFunc = cfg.ContainerFunc
	}

	return &KubeApplierExecutor{
		ExecutorBundle:   bundle,
		BundleName:       cfg.PhaseName,
		ExecutorDocument: cfg.ExecutorDocument,
		apiObject:        apiObj,
		clusterMap:       cfg.ClusterMap,
		clusterName:      cfg.ClusterName,
		kubeconfig:       cfg.KubeConfig,
		clientFunc:       clientFunc,
		execObj:          cObj,
		targetPath:       cfg.TargetPath,
	}, nil
}

// Run executor, should be performed in separate go routine
func (e *KubeApplierExecutor) Run(runOpts ifc.RunOptions) error {
	e.apiObject.Config.Debug = log.DebugEnabled()
	e.apiObject.Config.PhaseName = e.BundleName

	if e.apiObject.Config.Kubeconfig == "" {
		kcfg, ctx, cleanup, err := e.getKubeconfig()
		if err != nil {
			return err
		}
		defer cleanup()
		e.apiObject.Config.Kubeconfig, e.apiObject.Config.Context = kcfg, ctx
	}
	e.execObj.Spec.StorageMounts = append(e.execObj.Spec.StorageMounts, airshipv1.StorageMount{
		MountType:     "bind",
		Src:           e.apiObject.Config.Kubeconfig,
		DstPath:       e.apiObject.Config.Kubeconfig,
		ReadWriteMode: false,
	})
	log.Printf("using kubeconfig at '%s' and context '%s'", e.apiObject.Config.Kubeconfig, e.apiObject.Config.Context)

	e.apiObject.Config.DryRun = runOpts.DryRun
	if runOpts.Timeout != nil {
		e.apiObject.Config.WaitOptions.Timeout = int(*runOpts.Timeout)
	}

	reader, err := e.prepareDocuments()
	if err != nil {
		return err
	}

	opts, err := yaml.Marshal(&e.apiObject.Config)
	if err != nil {
		return err
	}

	e.execObj.Config = string(opts)
	return e.clientFunc("", reader, os.Stdout, e.execObj, e.targetPath).Run()
}

func (e *KubeApplierExecutor) getKubeconfig() (string, string, func(), error) {
	log.Debug("Getting kubeconfig context name from cluster map")
	ctx, err := e.clusterMap.ClusterKubeconfigContext(e.clusterName)
	if err != nil {
		return "", "", nil, err
	}
	log.Debug("Getting kubeconfig file information from kubeconfig provider")
	path, cleanup, err := e.kubeconfig.GetFile()
	if err != nil {
		return "", "", nil, err
	}
	return path, ctx, cleanup, nil
}

func (e *KubeApplierExecutor) prepareDocuments() (io.Reader, error) {
	log.Debug("Filtering out documents that shouldn't be applied to kubernetes from document bundle")
	filteredBundle, err := e.ExecutorBundle.SelectBundle(document.NewDeployToK8sSelector())
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	err = filteredBundle.Write(buf)
	if err != nil {
		return nil, err
	}
	return buf, err
}

// Validate document set
func (e *KubeApplierExecutor) Validate() error {
	if e.BundleName == "" {
		return errors.ErrInvalidPhase{Reason: "k8s applier BundleName is empty"}
	}
	docs, err := e.ExecutorBundle.GetAllDocuments()
	if err != nil {
		return err
	}
	if len(docs) == 0 {
		return errors.ErrInvalidPhase{Reason: "no executor documents in the bundle"}
	}
	// TODO: need to find if any other validation needs to be added
	return nil
}

// Render document set
func (e *KubeApplierExecutor) Render(w io.Writer, o ifc.RenderOptions) error {
	bundle, err := e.ExecutorBundle.SelectBundle(o.FilterSelector)
	if err != nil {
		return err
	}
	return bundle.Write(w)
}

// Status returns the status of the given phase
func (e *KubeApplierExecutor) Status() (ifc.ExecutorStatus, error) {
	return ifc.ExecutorStatus{}, airerrors.ErrNotImplemented{What: KubernetesApply}
}
