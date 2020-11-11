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

package document_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"bufio"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/fs"
)

func KustomNodeTestdata(writer io.Writer) document.KustomNode {
	return document.KustomNode{
		Name: "workers-targetphase/kustomization.yaml",
		Data: "testdata/workers-targetphase/kustomization.yaml",
		Children: []document.KustomNode{
			{
				Name: "Resources",
				Data: "",
				Children: []document.KustomNode{
					{
						Name:     "workers-targetphase/nodes/kustomization.yaml",
						Data:     "testdata/workers-targetphase/nodes/kustomization.yaml",
						Children: []document.KustomNode{},
					},
				},
				Writer: writer,
			},
			{
				Name: "Generators",
				Data: "",
				Children: []document.KustomNode{
					{
						Name: "workers-targetphase/hostgenerator/kustomization.yaml",
						Data: "testdata/workers-targetphase/hostgenerator/kustomization.yaml",
						Children: []document.KustomNode{{
							Name: "Resources",
							Data: "",
							Children: []document.KustomNode{{
								Name: "workers-targetphase/hostgenerator/host-generation.yaml",
								Data: "testdata/workers-targetphase/hostgenerator/host-generation.yaml",
							},
							},
							Writer: writer,
						},
						},
					},
				},
				Writer: writer,
			},
		},
		Writer: writer,
	}
}

func TestBuildKustomTree(t *testing.T) {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	type args struct {
		entrypoint string
	}
	tests := []struct {
		name        string
		args        func(t *testing.T) args
		want1       document.KustomNode
		errContains string
	}{
		{
			name: "success build tree",
			args: func(t *testing.T) args {
				return args{entrypoint: "testdata/workers-targetphase/kustomization.yaml"}
			},
			want1: KustomNodeTestdata(w),
		},
		{
			name: "entrypoint doesn't exist",
			args: func(t *testing.T) args {
				return args{entrypoint: "tdata/kustomization.yaml"}
			},
			want1:       KustomNodeTestdata(w),
			errContains: "no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)
			manifestsDir := "testdata"
			got1, actualErr := document.BuildKustomTree(tArgs.entrypoint, w, manifestsDir)
			if tt.errContains != "" {
				require.Error(t, actualErr)
				assert.Contains(t, actualErr.Error(), tt.errContains)
			} else {
				require.NoError(t, actualErr)
				assert.Equal(t, got1.Name, tt.want1.Name)
				assert.Equal(t, len(got1.Children), len(tt.want1.Children))
			}
		})
	}
}

func Test_makeResMap(t *testing.T) {
	type args struct {
		kfile string
		fs    fs.FileSystem
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		want1       map[string][]string
		errContains string
	}{
		{
			args: func(t *testing.T) args {
				return args{kfile: "testdata/workers-targetphase/kustomization.yaml", fs: fs.NewDocumentFs()}
			},
			name: "success resmap",
			want1: map[string][]string{
				"Generators": {
					"testdata/workers-targetphase/hostgenerator",
				},
				"Resources": {
					"testdata/workers-targetphase/nodes",
				},
			},
		},
		{
			args: func(t *testing.T) args {
				return args{kfile: "testdata/no_plan_site/phases/kustomization.yaml"}
			},
			name:        "nil case",
			want1:       map[string][]string{},
			errContains: "received nil filesystem",
		},
		{
			args: func(t *testing.T) args {
				return args{kfile: "t/workers-targetphase/kustomization.yaml", fs: fs.NewDocumentFs()}
			},
			name: "fail resmap,entrypoint not found",
			want1: map[string][]string{
				"Resources": {
					"testdata/workers-targetphase/nodes",
				},
			},
			errContains: "no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got1, actualErr := document.MakeResMap(tArgs.fs, tArgs.kfile)
			if tt.errContains != "" {
				require.Error(t, actualErr)
				assert.Contains(t, actualErr.Error(), tt.errContains)
			} else {
				require.NoError(t, actualErr)
				assert.Equal(t, got1, tt.want1)
			}
		})
	}
}

func TestKustomNode_PrintTree(t *testing.T) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	type args struct {
		prefix string
	}
	tests := []struct {
		name string
		init func(t *testing.T) document.KustomNode
		want string

		args func(t *testing.T) args
	}{
		{
			name: "valid print tree",
			args: func(t *testing.T) args {
				return args{prefix: ""}
			},
			init: func(t *testing.T) document.KustomNode {
				return KustomNodeTestdata(writer)
			},
			want: "    └── hostgenerator [workers-targetphase/hostgenerator]\n",
		},
	}
	rescueStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		require.Error(t, err)
	}
	os.Stdout = w

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			receiver.PrintTree(tArgs.prefix)
			w.Close()
			out, err := ioutil.ReadAll(r)
			if err != nil {
				require.Error(t, err)
			}
			os.Stdout = rescueStdout
			assert.Equal(t, string(out), tt.want)
		})
	}
}

func Test_getKustomChildren(t *testing.T) {
	type args struct {
		k document.KustomNode
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		want1 []document.KustomNode
	}{
		{
			name: "success getkustomchildren",
			args: func(t *testing.T) args {
				return args{k: document.KustomNode{
					Name: "Generators",
					Data: "",
					Children: []document.KustomNode{
						{
							Name: "workers-targetphase/hostgenerator/kustomization.yaml",
							Data: "testdata/workers-targetphase/hostgenerator/kustomization.yaml",
							Children: []document.KustomNode{{
								Name: "workers-targetphase/hostgenerator/host-generation.yaml",
								Data: "testdata/workers-targetphase/hostgenerator/host-generation.yaml",
							},
							},
						},
					},
				},
				}
			},
			want1: []document.KustomNode{
				{
					Name: "workers-targetphase/hostgenerator/kustomization.yaml",
					Data: "testdata/workers-targetphase/hostgenerator/kustomization.yaml",
					Children: []document.KustomNode{{
						Name: "workers-targetphase/hostgenerator/host-generation.yaml",
						Data: "testdata/workers-targetphase/hostgenerator/host-generation.yaml",
					},
					},
				},
			},
		},
		{
			name: "no children nodes",
			args: func(t *testing.T) args {
				return args{k: document.KustomNode{
					Name: "Transformers",
					Data: "",
				}}
			},
			want1: []document.KustomNode{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			got1 := document.GetKustomChildren(tArgs.k)
			assert.Equal(t, got1, tt.want1)
		})
	}
}
