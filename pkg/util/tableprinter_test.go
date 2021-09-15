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

package util_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/util"
)

func TestNewTabPrinter(t *testing.T) {
	var tests = []struct {
		name      string
		objects   interface{}
		template  string
		noHeaders bool
		expected  string
		errorStr  string
	}{
		{
			name:     "nil object",
			template: "NAME:name",
		},
		{
			name:     "no template defined",
			errorStr: "custom-columns format specified but no custom columns given",
		},
		{
			name: "success single object",
			objects: &v1alpha1.Phase{
				ObjectMeta: metav1.ObjectMeta{Name: "phase"},
			},
			template: "NAME:metadata.name",
			expected: "NAME\nphase\n",
		},
		{
			name: "success multiple objects",
			objects: []*v1alpha1.Phase{{
				ObjectMeta: metav1.ObjectMeta{Name: "phase1"},
			},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "phase2"},
				},
			},
			template: "NAME:metadata.name",
			expected: "NAME\nphase1\nphase2\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := util.PrintObjects(tt.objects, tt.template, buf, tt.noHeaders)
			if tt.errorStr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorStr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, buf.String())
			}
			buf.Reset()
		})
	}
}
