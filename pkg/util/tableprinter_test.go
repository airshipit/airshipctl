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
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	airapiv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/util"
)

func TestPrintTableForList(t *testing.T) {
	resources := []*airapiv1.Phase{
		{
			TypeMeta: metav1.TypeMeta{
				Kind: "Phase",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "p1",
			},
		},
	}
	expectedOut := [][]byte{
		[]byte("NAMESPACE   RESOURCE                                "),
		[]byte("            Phase/p1                                "),
		{},
	}

	rt, err := util.NewResourceTable(resources, func(util.Printable) *util.PrintResourceStatus { return nil })
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	util.DefaultTablePrinter(buf, nil).PrintTable(rt, 0)
	out, err := ioutil.ReadAll(buf)
	require.NoError(t, err)
	assert.Equal(t, expectedOut, bytes.Split(out, []byte("\n")))
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
			"executorRef":        nil,
		},
	}
	printable := &airapiv1.Phase{
		TypeMeta: metav1.TypeMeta{
			Kind: "Phase",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "p1",
		},
	}
	rs := f(printable)
	assert.Equal(t, expectedObj, rs.Resource.Object)
}

func TestNonPrintable(t *testing.T) {
	_, err := util.NewResourceTable("non Printable string", util.DefaultStatusFunction())
	assert.Error(t, err)
}

func TestPrintTableForSingleResource(t *testing.T) {
	resource := &airapiv1.Phase{
		TypeMeta: metav1.TypeMeta{
			Kind: "Phase",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "p1",
		},
	}
	expectedOut := [][]byte{
		[]byte("NAMESPACE   RESOURCE                                "),
		[]byte("            Phase/p1                                "),
		{},
	}
	rt, err := util.NewResourceTable(resource, func(util.Printable) *util.PrintResourceStatus { return nil })
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	util.DefaultTablePrinter(buf, nil).PrintTable(rt, 0)
	out, err := ioutil.ReadAll(buf)
	require.NoError(t, err)
	assert.Equal(t, expectedOut, bytes.Split(out, []byte("\n")))
}
