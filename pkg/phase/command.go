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

package phase

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

// RunFlags options for phase run command
type RunFlags struct {
	DryRun     bool
	Timeout    time.Duration
	PhaseID    ifc.ID
	Kubeconfig string
	Progress   bool
}

// RunCommand phase run command
type RunCommand struct {
	Options RunFlags
	Factory config.Factory
}

// RunE runs the phase
func (c *RunCommand) RunE() error {
	cfg, err := c.Factory()
	if err != nil {
		return err
	}

	helper, err := NewHelper(cfg)
	if err != nil {
		return err
	}

	kubeconfigOption := InjectKubeconfigPath(c.Options.Kubeconfig)
	client := NewClient(helper, kubeconfigOption)

	phase, err := client.PhaseByID(c.Options.PhaseID)
	if err != nil {
		return err
	}
	return phase.Run(ifc.RunOptions{DryRun: c.Options.DryRun, Timeout: c.Options.Timeout, Progress: c.Options.Progress})
}

// PlanCommand plan command
type PlanCommand struct {
	Factory config.Factory
	Writer  io.Writer
}

// RunE runs a phase plan command
func (c *PlanCommand) RunE() error {
	cfg, err := c.Factory()
	if err != nil {
		return err
	}

	helper, err := NewHelper(cfg)
	if err != nil {
		return err
	}

	plan, err := helper.Plan()
	if err != nil {
		return err
	}

	return PrintPlan(plan, c.Writer)
}

// TreeCommand plan command
type TreeCommand struct {
	Factory  config.Factory
	PhaseID  ifc.ID
	Writer   io.Writer
	Argument string
}

// RunE runs the phase tree command
func (c *TreeCommand) RunE() error {
	var entrypoint string
	cfg, err := c.Factory()
	if err != nil {
		return err
	}

	helper, err := NewHelper(cfg)
	if err != nil {
		return err
	}

	client := NewClient(helper)
	var manifestsDir string
	// check if its a relative path
	if _, err = os.Stat(c.Argument); err == nil {
		// capture manifests directory from phase relative path
		manifestsDir = strings.SplitAfter(c.Argument, "/manifests")[0]
		entrypoint = filepath.Join(c.Argument, document.KustomizationFile)
	} else {
		c.PhaseID.Name = c.Argument
		manifestsDir = filepath.Join(helper.TargetPath(), helper.PhaseRepoDir())
		var phase ifc.Phase
		phase, err = client.PhaseByID(c.PhaseID)
		if err != nil {
			return err
		}
		var rootPath string
		rootPath, err = phase.DocumentRoot()
		if err != nil {
			return err
		}
		entrypoint = filepath.Join(rootPath, document.KustomizationFile)
	}
	t, err := document.BuildKustomTree(entrypoint, c.Writer, manifestsDir)
	if err != nil {
		return err
	}
	t.PrintTree("")
	return nil
}
