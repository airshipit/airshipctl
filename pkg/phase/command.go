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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cli-utils/pkg/print/table"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
	"opendev.org/airship/airshipctl/pkg/util"
)

// GenericRunFlags generic options for run command
type GenericRunFlags struct {
	DryRun     bool
	Timeout    time.Duration
	Kubeconfig string
	Progress   bool
}

// RunFlags options for phase run command
type RunFlags struct {
	GenericRunFlags
	PhaseID ifc.ID
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

// ListCommand phase list command
type ListCommand struct {
	Factory config.Factory
	Writer  io.Writer
}

// RunE runs a phase plan command
func (c *ListCommand) RunE() error {
	cfg, err := c.Factory()
	if err != nil {
		return err
	}

	helper, err := NewHelper(cfg)
	if err != nil {
		return err
	}

	phases, err := helper.ListPhases()
	if err != nil {
		return err
	}

	rt, err := util.NewResourceTable(phases, util.DefaultStatusFunction())
	if err != nil {
		return err
	}

	util.DefaultTablePrinter(c.Writer, nil).PrintTable(rt, 0)
	return nil
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

// PlanListCommand phase list command
type PlanListCommand struct {
	Factory config.Factory
	Writer  io.Writer
}

// RunE runs a phase plan command
func (c *PlanListCommand) RunE() error {
	cfg, err := c.Factory()
	if err != nil {
		return err
	}

	helper, err := NewHelper(cfg)
	if err != nil {
		return err
	}

	phases, err := helper.ListPlans()
	if err != nil {
		return err
	}

	rt, err := util.NewResourceTable(phases, util.DefaultStatusFunction())
	if err != nil {
		return err
	}

	printer := util.DefaultTablePrinter(c.Writer, nil)
	descriptionCol := table.ColumnDef{
		ColumnName:   "description",
		ColumnHeader: "DESCRIPTION",
		ColumnWidth:  40,
		PrintResourceFunc: func(w io.Writer, width int, r table.Resource) (int, error) {
			rs := r.ResourceStatus()
			if rs == nil {
				return 0, nil
			}
			plan := &v1alpha1.PhasePlan{}
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(rs.Resource.Object, plan)
			if err != nil {
				return 0, err
			}
			return fmt.Fprint(w, plan.Description)
		},
	}
	printer.Columns = append(printer.Columns, descriptionCol)
	printer.PrintTable(rt, 0)
	return nil
}

// PlanRunFlags options for phase run command
type PlanRunFlags struct {
	GenericRunFlags
	PlanID ifc.ID
}

// PlanRunCommand phase run command
type PlanRunCommand struct {
	Options PlanRunFlags
	Factory config.Factory
}

// RunE executes phase plan
func (c *PlanRunCommand) RunE() error {
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

	plan, err := client.PlanByID(c.Options.PlanID)
	if err != nil {
		return err
	}
	return plan.Run(ifc.RunOptions{DryRun: c.Options.DryRun, Timeout: c.Options.Timeout, Progress: c.Options.Progress})
}
