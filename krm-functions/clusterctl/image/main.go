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
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	clusterctl       = "clusterctl"
	clusterAPIConfig = "clusterctl.yaml"

	dirPerm  = 0755
	filePerm = 0644
)

// ClusterctlOptions holds all necessary data to run clusterctl inside of KRM
type ClusterctlOptions struct {
	CmdOptions []string          `json:"cmd-options,omitempty"`
	Config     []byte            `json:"config,omitempty"`
	Components map[string][]byte `json:"components,omitempty"`
}

// Run prepares config, repo tree and executes clusterctl with appropriate options
func (c *ClusterctlOptions) Run([]*yaml.RNode) ([]*yaml.RNode, error) {
	if err := c.buildRepoTree(); err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(clusterAPIConfig, c.Config, filePerm); err != nil {
		return nil, err
	}

	return nil, runCmd(clusterctl, c.CmdOptions)
}

func (c *ClusterctlOptions) buildRepoTree() error {
	for f, component := range c.Components {
		componentDir := filepath.Dir(f)
		if _, err := os.Stat(componentDir); os.IsNotExist(err) {
			if err := os.MkdirAll(componentDir, dirPerm); err != nil {
				return err
			}
		}
		if err := ioutil.WriteFile(f, component, filePerm); err != nil {
			return err
		}
	}
	return nil
}

func runCmd(cmd string, opts []string) error {
	printMsg("#%s %s\n", cmd, strings.Join(opts, " "))
	c := exec.Command(cmd, opts...)
	// allows to observe realtime output from script
	w := io.Writer(os.Stderr)
	c.Stdout = w
	c.Stderr = w
	return c.Run()
}

// printMsg is a convenient function to print output to stderr
func printMsg(format string, a ...interface{}) {
	if _, err := fmt.Fprintf(os.Stderr, format, a...); err != nil {
	}
}

func main() {
	cfg := &ClusterctlOptions{}
	if err := command.Build(framework.SimpleProcessor{Filter: kio.FilterFunc(cfg.Run), Config: cfg},
	command.StandaloneDisabled, false).Execute(); err != nil {
		printMsg("\n")
		os.Exit(1)
	}
}
