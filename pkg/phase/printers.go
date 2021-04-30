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
	"errors"
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/runtime"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/util"

	"sigs.k8s.io/cli-utils/pkg/print/table"
)

//PrintPhaseListTable prints phase list table
func PrintPhaseListTable(w io.Writer, phases []*v1alpha1.Phase) error {
	rt, err := util.NewResourceTable(phases, util.DefaultStatusFunction())
	if err != nil {
		return err
	}
	printer := util.DefaultTablePrinter(w, nil)
	clusternameCol := table.ColumnDef{
		ColumnName:   "clustername",
		ColumnHeader: "CLUSTER NAME",
		ColumnWidth:  20,
		PrintResourceFunc: func(w io.Writer, width int, r table.Resource) (int,
			error) {
			phase, err := phaseFromResource(r)
			if err != nil {
				return 0, nil
			}
			txt := phase.ClusterName
			if len(txt) > width {
				txt = txt[:width]
			}
			_, err = fmt.Fprintf(w, txt)
			return len(txt), err
		},
	}
	executorrefkindCol := table.ColumnDef{
		ColumnName:   "executorrefkind",
		ColumnHeader: "EXECUTOR",
		ColumnWidth:  20,
		PrintResourceFunc: func(w io.Writer, width int, r table.Resource) (int,
			error) {
			phase, err := phaseFromResource(r)
			if err != nil {
				return 0, nil
			}
			txt := phase.Config.ExecutorRef.Kind
			if len(txt) > width {
				txt = txt[:width]
			}
			_, err = fmt.Fprintf(w, txt)
			return len(txt), err
		},
	}
	docentrypointCol := table.ColumnDef{
		ColumnName:   "docentrypoint",
		ColumnHeader: "DOC ENTRYPOINT",
		ColumnWidth:  100,
		PrintResourceFunc: func(w io.Writer, width int, r table.Resource) (int,
			error) {
			phase, err := phaseFromResource(r)
			if err != nil {
				return 0, nil
			}
			txt := phase.Config.DocumentEntryPoint
			if len(txt) > width {
				txt = txt[:width]
			}
			_, err = fmt.Fprintf(w, txt)
			return len(txt), err
		},
	}
	printer.Columns = append(printer.Columns, clusternameCol, executorrefkindCol, docentrypointCol)
	printer.PrintTable(rt, 0)
	return nil
}

func phaseFromResource(r table.Resource) (*v1alpha1.Phase, error) {
	rs := r.ResourceStatus()
	if rs == nil {
		return nil, errors.New("Resource status is nil")
	}
	phase := &v1alpha1.Phase{}
	return phase, runtime.DefaultUnstructuredConverter.FromUnstructured(rs.Resource.Object, phase)
}

//PrintPlanListTable prints plan list table
func PrintPlanListTable(w io.Writer, phasePlans []*v1alpha1.PhasePlan) error {
	rt, err := util.NewResourceTable(phasePlans, util.DefaultStatusFunction())
	if err != nil {
		return err
	}

	printer := util.DefaultTablePrinter(w, nil)
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
			_, err = fmt.Fprintf(w, txt)
			return len(txt), err
		},
	}
	printer.Columns = append(printer.Columns, descriptionCol)
	printer.PrintTable(rt, 0)
	return nil
}
