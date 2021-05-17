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
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	// TODO this small library needs to be moved to airshipctl and extended
	// with splitting streams into Stderr and Stdout
	"github.com/ahmetb/dlog"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/runfn"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/util"
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
	conf *v1alpha1.GenericContainer,
	targetPath string) ClientV1Alpha1

type clientV1Alpha1 struct {
	resultsDir string
	input      io.Reader
	output     io.Writer
	conf       *v1alpha1.GenericContainer
	targetPath string

	containerFunc containerFunc
}

type containerFunc func(ctx context.Context, driver string, url string) (Container, error)

// NewClientV1Alpha1 constructor for ClientV1Alpha1
func NewClientV1Alpha1(
	resultsDir string,
	input io.Reader,
	output io.Writer,
	conf *v1alpha1.GenericContainer,
	targetPath string) ClientV1Alpha1 {
	return &clientV1Alpha1{
		resultsDir:    resultsDir,
		output:        output,
		input:         input,
		conf:          conf,
		containerFunc: NewContainer,
		targetPath:    targetPath,
	}
}

// Run will perform container run action based on the configuration
func (c *clientV1Alpha1) Run() error {
	// expand Src paths for mount if they are relative
	ExpandSourceMounts(c.conf.Spec.StorageMounts, c.targetPath)
	// set default runtime
	switch c.conf.Spec.Type {
	case v1alpha1.GenericContainerTypeAirship, "":
		return c.runAirship()
	case v1alpha1.GenericContainerTypeKrm:
		return c.runKRM()
	default:
		return fmt.Errorf("unknown generic container type %s", c.conf.Spec.Type)
	}
}

func (c *clientV1Alpha1) runAirship() error {
	if c.conf.Spec.Airship.ContainerRuntime == "" {
		c.conf.Spec.Airship.ContainerRuntime = ContainerDriverDocker
	}

	var cont Container
	if c.containerFunc == nil {
		c.containerFunc = NewContainer
	}

	cont, err := c.containerFunc(
		context.Background(),
		c.conf.Spec.Airship.ContainerRuntime,
		c.conf.Spec.Image)
	if err != nil {
		return err
	}
	defer func(container Container) {
		if rmErr := container.RmContainer(); rmErr != nil {
			log.Printf("Failed to remove container with id '%s', err is '%s'", container.GetID(), rmErr.Error())
		}
	}(cont)

	// this will split the env vars into the ones to be exported and the ones that have values
	contEnv := runtimeutil.NewContainerEnvFromStringSlice(c.conf.Spec.EnvVars)

	envs := []string{}
	for _, key := range contEnv.VarsToExport {
		envs = append(envs, strings.Join([]string{key, os.Getenv(key)}, "="))
	}

	for key, value := range contEnv.EnvVars {
		envs = append(envs, strings.Join([]string{key, value}, "="))
	}

	node, err := kyaml.Parse(c.conf.Config)
	if err != nil {
		return err
	}

	decoratedInput := bytes.NewBuffer([]byte{})
	pipeline := &kio.Pipeline{
		Inputs: []kio.Reader{&kio.ByteReader{Reader: c.input}},
		Outputs: []kio.Writer{kio.ByteWriter{
			Writer:                decoratedInput,
			KeepReaderAnnotations: true,
			WrappingKind:          kio.ResourceListKind,
			WrappingAPIVersion:    kio.ResourceListAPIVersion,
			FunctionConfig:        node,
		}},
	}

	err = pipeline.Execute()
	if err != nil {
		return err
	}

	log.Printf("Starting container with image: '%s', cmd: '%s'",
		c.conf.Spec.Image,
		c.conf.Spec.Airship.Cmd)
	err = cont.RunCommand(RunCommandOptions{
		Privileged:  c.conf.Spec.Airship.Privileged,
		Cmd:         c.conf.Spec.Airship.Cmd,
		Mounts:      convertDockerMount(c.conf.Spec.StorageMounts),
		EnvVars:     envs,
		Input:       decoratedInput,
		HostNetwork: c.conf.Spec.HostNetwork,
	})
	if err != nil {
		return err
	}

	log.Debugf("Waiting for container run to finish, image: '%s', cmd: '%s'",
		c.conf.Spec.Image,
		c.conf.Spec.Airship.Cmd)

	// write logs asynchronously while waiting for for container to finish
	cErr := make(chan error, 1)
	go func() {
		cErr <- writeLogs(cont)
	}()

	err = cont.WaitUntilFinished()
	if err != nil {
		<-cErr
		return err
	}

	// check writeLogs error after container is done waiting
	if err = <-cErr; err != nil {
		return err
	}

	rOut, err := cont.GetContainerLogs(GetLogOptions{Stdout: true})
	if err != nil {
		return err
	}
	defer rOut.Close()

	parsedOut := dlog.NewReader(rOut)

	return writeSink(c.resultsDir, parsedOut, c.output)
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

func writeLogs(cont Container) error {
	stderr, err := cont.GetContainerLogs(GetLogOptions{
		Stderr: true,
		Follow: true})
	if err != nil {
		return err
	}
	defer stderr.Close()
	parsedStdErr := dlog.NewReader(stderr)
	_, err = io.Copy(log.Writer(), parsedStdErr)
	return err
}

// writeSink output to directory on filesystem sink
func writeSink(path string, rc io.Reader, out io.Writer) error {
	inputs := []kio.Reader{&kio.ByteReader{Reader: rc}}
	var outputs []kio.Writer
	switch {
	case out == nil && path != "":
		log.Debugf("writing container output to files in directory %s", path)
		outputs = []kio.Writer{&kio.LocalPackageWriter{PackagePath: path}}
	case out != nil:
		log.Debugf("writing container output to provided writer")
		outputs = []kio.Writer{&kio.ByteWriter{Writer: out}}
	default:
		log.Debugf("writing container output to stdout")
		outputs = []kio.Writer{&kio.ByteWriter{Writer: os.Stdout}}
	}
	return kio.Pipeline{Inputs: inputs, Outputs: outputs}.Execute()
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

func convertDockerMount(airMounts []v1alpha1.StorageMount) (mounts []Mount) {
	for _, mount := range airMounts {
		mnt := Mount{
			Type: mount.MountType,
			Src:  mount.Src,
			Dst:  mount.DstPath,
		}
		if !mount.ReadWriteMode {
			mnt.ReadOnly = true
		}
		mounts = append(mounts, mnt)
	}
	return mounts
}

// ExpandSourceMounts converts relative paths into absolute ones
func ExpandSourceMounts(storageMounts []v1alpha1.StorageMount, targetPath string) {
	for i, mount := range storageMounts {
		// Try to expand Src path
		path := util.ExpandTilde(mount.Src)
		// If still relative - add targetPath prefix
		if !filepath.IsAbs(path) {
			path = filepath.Join(targetPath, mount.Src)
		}
		storageMounts[i].Src = path
	}
}
