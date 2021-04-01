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

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	kerror "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"

	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	// EnvRenderedBundlePath will be passed to the script, it will contain path to the rendered bundle
	EnvRenderedBundlePath = "RENDERED_BUNDLE_PATH"
	scriptPath            = "script.sh"
	scriptKey             = "script"
	bundleFile            = "bundle.yaml"
	workdir               = "/tmp"
)

func main() {
	cfg := &v1.ConfigMap{}
	resourceList := &framework.ResourceList{FunctionConfig: &cfg}
	runner := ScriptRunner{
		ScriptFile:         scriptPath,
		WorkDir:            workdir,
		RenderedBundleFile: bundleFile,
		DataKey:            scriptKey,
		ResourceList:       resourceList,
		ConfigMap:          cfg,
		ErrStream:          os.Stderr,
		OutStream:          os.Stdout,
	}
	cmd := framework.Command(resourceList, runner.Run)
	if err := cmd.Execute(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

// ScriptRunner writes to file system and executes the script
type ScriptRunner struct {
	ScriptFile, WorkDir, DataKey, RenderedBundleFile string

	ErrStream io.Writer
	OutStream io.Writer

	ConfigMap    *v1.ConfigMap
	ResourceList *framework.ResourceList
}

// Run writes the script and bundle to the file system and executes it
func (c *ScriptRunner) Run() error {
	bundlePath, scriptPath := c.getBundleAndScriptPath()

	script, exist := c.ConfigMap.Data[c.DataKey]
	if !exist {
		return fmt.Errorf("ConfigMap '%s/%s' doesnt' have specified script key '%s'",
			c.ConfigMap.Namespace, c.ConfigMap.Name, c.DataKey)
	}

	err := ioutil.WriteFile(scriptPath, []byte(script), 0555)
	if err != nil {
		return err
	}
	err = c.writeBundle(bundlePath, c.ResourceList.Items)
	if err != nil {
		return err
	}

	c.ResourceList.Items = nil

	clicmd := exec.Command(scriptPath)
	clicmd.Stdout = c.OutStream
	clicmd.Stderr = c.ErrStream

	clicmd.Env = os.Environ()
	clicmd.Env = append(clicmd.Env, fmt.Sprintf("%s=%s", EnvRenderedBundlePath, bundlePath))

	err = clicmd.Start()
	if err != nil {
		return err
	}

	return clicmd.Wait()
}

// Cleanup removes script and bundle files from filesystem
func (c *ScriptRunner) Cleanup() error {
	bundlePath, scriptPath := c.getBundleAndScriptPath()

	scriptErr := os.Remove(scriptPath)
	if os.IsNotExist(scriptErr) {
		// If file doesn't exist no error happened
		scriptErr = nil
	}

	bundleErr := os.Remove(bundlePath)
	if os.IsNotExist(bundleErr) {
		// If file doesn't exist no error happened
		bundleErr = nil
	}

	return kerror.NewAggregate([]error{scriptErr, bundleErr})
}

func (c *ScriptRunner) getBundleAndScriptPath() (string, string) {
	return filepath.Join(c.WorkDir, c.RenderedBundleFile), filepath.Join(c.WorkDir, c.ScriptFile)
}

func (c *ScriptRunner) writeBundle(path string, items []*kyaml.RNode) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	pipeline := kio.Pipeline{
		Outputs: []kio.Writer{
			kio.ByteWriter{
				Writer: f,
			},
		},
		Inputs: []kio.Reader{
			kio.ResourceNodeSlice(items),
		},
	}

	return pipeline.Execute()
}
