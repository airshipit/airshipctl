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
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/util"
)

func TestPrintPhaseListTable(t *testing.T) {
	phases := []*v1alpha1.Phase{
		{
			TypeMeta: metav1.TypeMeta{
				Kind: "Phase",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        "p1",
				ClusterName: "cluster",
			},
			Config: v1alpha1.PhaseConfig{
				DocumentEntryPoint: "test",
				ExecutorRef:        &v1.ObjectReference{Kind: "test"},
			},
		},
	}

	tests := []struct {
		name      string
		phases    []*v1alpha1.Phase
		wantPanic bool
	}{
		{
			name:      "success",
			phases:    phases,
			wantPanic: false,
		},
		{
			name: "phase with no executor ref",
			phases: []*v1alpha1.Phase{
				{
					TypeMeta: metav1.TypeMeta{
						Kind: "Pe",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "p1",
					},
				},
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("should have panic")
				}
			}()
			err := PrintPhaseListTable(w, tt.phases)
			require.NoError(t, err)
		})
	}
}

func TestNonPrintable(t *testing.T) {
	_, err := util.NewResourceTable("non Printable string", util.DefaultStatusFunction())
	assert.Error(t, err)
}

func TestDefaultStatusFunction(t *testing.T) {
	f := util.DefaultStatusFunction()
	expectedObj := map[string]interface{}{
		"kind": "Phase",
		"metadata": map[string]interface{}{
			"name":              "p1",
			"creationTimestamp": nil,
		},
		"config": map[string]interface{}{
			"documentEntryPoint": "",
			"executorRef": map[string]interface{}{
				"kind": "test",
			},
			"validation": map[string]interface{}{},
		},
	}
	printable := &v1alpha1.Phase{
		TypeMeta: metav1.TypeMeta{
			Kind: "Phase",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "p1",
		},
		Config: v1alpha1.PhaseConfig{
			ExecutorRef: &v1.ObjectReference{Kind: "test"},
		},
	}
	rs := f(printable)
	assert.Equal(t, expectedObj, rs.Resource.Object)
}

func TestPrintPlanListTable(t *testing.T) {
	plans := []*v1alpha1.PhasePlan{
		{
			TypeMeta: metav1.TypeMeta{
				Kind: "PhasePlan",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "p",
			},
			Description: "description",
			Phases: []v1alpha1.PhaseStep{
				{
					Name: "phase",
				},
			},
		},
	}
	tests := []struct {
		name  string
		plans []*v1alpha1.PhasePlan
	}{
		{
			name:  "Success print plan list",
			plans: plans,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			err := PrintPlanListTable(w, tt.plans)
			require.NoError(t, err)
		})
	}
}

func TestDefaultStatusFunctionForPhasePlan(t *testing.T) {
	f := util.DefaultStatusFunction()
	expectedObj := map[string]interface{}{
		"kind": "PhasePlan",
		"metadata": map[string]interface{}{
			"name":              "p1",
			"creationTimestamp": nil,
		},
		"validation": map[string]interface{}{},
	}
	printable := &v1alpha1.PhasePlan{
		TypeMeta: metav1.TypeMeta{
			Kind: "PhasePlan",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "p1",
		},
	}
	rs := f(printable)
	assert.Equal(t, expectedObj, rs.Resource.Object)
}
