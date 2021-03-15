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
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	phaseerrors "opendev.org/airship/airshipctl/pkg/phase/errors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
	"opendev.org/airship/airshipctl/pkg/util"
	"opendev.org/airship/airshipctl/pkg/util/yaml"
)

// GenericRunFlags generic options for run command
type GenericRunFlags struct {
	DryRun  bool
	Timeout time.Duration
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

	client := NewClient(helper)

	phase, err := client.PhaseByID(c.Options.PhaseID)
	if err != nil {
		return err
	}
	return phase.Run(ifc.RunOptions{DryRun: c.Options.DryRun, Timeout: c.Options.Timeout})
}

// ListCommand phase list command
type ListCommand struct {
	Factory      config.Factory
	Writer       io.Writer
	ClusterName  string
	PlanID       ifc.ID
	OutputFormat string
}

// RunE runs a phase list command
func (c *ListCommand) RunE() error {
	if c.OutputFormat != "table" && c.OutputFormat != "yaml" {
		return phaseerrors.ErrInvalidFormat{RequestedFormat: c.OutputFormat}
	}
	cfg, err := c.Factory()
	if err != nil {
		return err
	}

	helper, err := NewHelper(cfg)
	if err != nil {
		return err
	}

	o := ifc.ListPhaseOptions{ClusterName: c.ClusterName, PlanID: c.PlanID}
	phaseList, err := helper.ListPhases(o)
	if err != nil {
		return err
	}
	if c.OutputFormat == "table" {
		return PrintPhaseListTable(c.Writer, phaseList)
	}
	return yaml.WriteOut(c.Writer, phaseList)
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

// RunE runs a plan list command
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
		ColumnWidth:  200,
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
			txt := plan.Description
			if len(txt) > width {
				txt = txt[:width]
			}
			_, err = fmt.Fprint(w, txt)
			return len(txt), err
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

	client := NewClient(helper)

	plan, err := client.PlanByID(c.Options.PlanID)
	if err != nil {
		return err
	}
	return plan.Run(ifc.RunOptions{DryRun: c.Options.DryRun, Timeout: c.Options.Timeout})
}

// ClusterListCommand options for cluster list command
type ClusterListCommand struct {
	Factory config.Factory
	Writer  io.Writer
	Format  string
}

// RunE executes cluster list command
func (c *ClusterListCommand) RunE() error {
	if c.Format != "table" && c.Format != "name" {
		return phaseerrors.ErrInvalidOutputFormat{RequestedFormat: c.Format}
	}
	cfg, err := c.Factory()
	if err != nil {
		return err
	}
	helper, err := NewHelper(cfg)
	if err != nil {
		return err
	}
	clusterMap, err := helper.ClusterMap()
	if err != nil {
		return err
	}
	err = clusterMap.Write(c.Writer, clustermap.WriteOptions{Format: c.Format})
	if err != nil {
		return err
	}
	return nil
}

// ValidateFlags options for phase validate command
type ValidateFlags struct {
	PhaseID ifc.ID
}

// ValidateCommand phase validate command
type ValidateCommand struct {
	Options ValidateFlags
	Factory config.Factory
}

// RunE runs the phase validate command
func (c *ValidateCommand) RunE() error {
	cfg, err := c.Factory()
	if err != nil {
		return err
	}

	util.Setenv(util.EnvVar{Key: "AIRSHIPCTL_CURRENT_PHASE", Value: c.Options.PhaseID.Name})
	helper, err := NewHelper(cfg)
	if err != nil {
		return err
	}

	client := NewClient(helper)

	phase, err := client.PhaseByID(c.Options.PhaseID)
	if err != nil {
		return err
	}
	return phase.Validate()
}

// StatusFlags is a struct to define status type
type StatusFlags struct {
	Timeout  time.Duration
	PhaseID  ifc.ID
	Progress bool
}

// StatusCommand is a struct which defines status
type StatusCommand struct {
	Options StatusFlags
	Factory config.Factory
}

// RunE returns the status of the given phase
func (s *StatusCommand) RunE() error {
	cfg, err := s.Factory()
	if err != nil {
		return err
	}

	helper, err := NewHelper(cfg)
	if err != nil {
		return err
	}

	ph, err := NewClient(helper).PhaseByID(s.Options.PhaseID)
	if err != nil {
		return err
	}

	_, err = ph.Status()
	return err
}

// PlanValidateFlags options for plan validate command
type PlanValidateFlags struct {
	PlanID ifc.ID
}

// PlanValidateCommand plan validate command
type PlanValidateCommand struct {
	Options PlanValidateFlags
	Factory config.Factory
}

// RunE runs the plan validate command
func (c *PlanValidateCommand) RunE() error {
	cfg, err := c.Factory()
	if err != nil {
		return err
	}

	util.Setenv(util.EnvVar{Key: "AIRSHIPCTL_CURRENT_PLAN", Value: c.Options.PlanID.Name})
	helper, err := NewHelper(cfg)
	if err != nil {
		return err
	}

	client := NewClient(helper)

	plan, err := client.PlanByID(c.Options.PlanID)
	if err != nil {
		return err
	}
	return plan.Validate()
}
