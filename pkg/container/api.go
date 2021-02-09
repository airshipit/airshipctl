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

package container

import (
	"context"
	"fmt"
	"io"

	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/runfn"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"

	"opendev.org/airship/airshipctl/pkg/errors"
)

// ClientV1Alpha1 provides airship generic container API
// TODO add generic mock for this client
type ClientV1Alpha1 interface {
	Run() error
}

// ClientV1Alpha1FactoryFunc used for tests
type ClientV1Alpha1FactoryFunc func(
	resultsDir string,
	input io.Reader,
	output io.Writer,
	conf *v1alpha1.GenericContainer) ClientV1Alpha1

type clientV1Alpha1 struct {
	resultsDir string
	input      io.Reader
	output     io.Writer
	conf       *v1alpha1.GenericContainer

	containerFunc containerFunc
}

type containerFunc func(ctx context.Context, driver string, url string) (Container, error)

// NewClientV1Alpha1 constructor for ClientV1Alpha1
func NewClientV1Alpha1(
	resultsDir string,
	input io.Reader,
	output io.Writer,
	conf *v1alpha1.GenericContainer) ClientV1Alpha1 {
	return &clientV1Alpha1{
		resultsDir:    resultsDir,
		output:        output,
		input:         input,
		conf:          conf,
		containerFunc: NewContainer,
	}
}

// Run will peform container run action based on the configuration
func (c *clientV1Alpha1) Run() error {
	// set default runtime
	switch c.conf.Spec.Type {
	case v1alpha1.GenericContainerTypeAirship, "":
		return errors.ErrNotImplemented{What: "airship generic container type"}
	case v1alpha1.GenericContainerTypeKrm:
		return c.runKRM()
	default:
		return fmt.Errorf("uknown generic container type %s", c.conf.Spec.Type)
	}
}

func (c *clientV1Alpha1) runKRM() error {
	mounts := convertKRMMount(c.conf.Spec.StorageMounts)
	fns := &runfn.RunFns{
		Network:               c.conf.Spec.HostNetwork,
		AsCurrentUser:         true,
		Path:                  c.resultsDir,
		Input:                 c.input,
		Output:                c.output,
		StorageMounts:         mounts,
		ContinueOnEmptyResult: true,
	}
	function, err := kyaml.Parse(c.conf.Config)
	if err != nil {
		return err
	}
	// Transform GenericContainer.Spec to annotation,
	// because we need to specify runFns config in annotation
	spec, err := yaml.Marshal(runtimeutil.FunctionSpec{
		Container: runtimeutil.ContainerSpec{
			Image:         c.conf.Spec.Image,
			Network:       c.conf.Spec.HostNetwork,
			Env:           c.conf.Spec.EnvVars,
			StorageMounts: mounts,
		},
	})
	if err != nil {
		return err
	}
	annotation := kyaml.SetAnnotation(runtimeutil.FunctionAnnotationKey, string(spec))
	_, err = annotation.Filter(function)
	if err != nil {
		return err
	}

	fns.Functions = []*kyaml.RNode{function}

	return fns.Execute()
}

func convertKRMMount(airMounts []v1alpha1.StorageMount) (fnsMounts []runtimeutil.StorageMount) {
	for _, mount := range airMounts {
		fnsMounts = append(fnsMounts, runtimeutil.StorageMount{
			MountType:     mount.MountType,
			Src:           mount.Src,
			DstPath:       mount.DstPath,
			ReadWriteMode: mount.ReadWriteMode,
		})
	}
	return fnsMounts
}
