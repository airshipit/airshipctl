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

package utils

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"opendev.org/airship/airshipctl/pkg/document"
)

func TestDefaultManifestFactory(t *testing.T) {
	bundle, err := document.NewBundleByPath("testdata/source_bundle")
	require.NoError(t, err)
	reader := DefaultManifestReaderFactory(false, bundle, FactoryFromKubeConfigPath("testdata/kubeconfig.yaml"))
	require.NotNil(t, reader)
}

func TestManifestBundleReader(t *testing.T) {
	bundle, err := document.NewBundleByPath("testdata/source_bundle")
	require.NoError(t, err)
	tests := []struct {
		name      string
		errString string

		reader io.Reader
		writer io.Writer
	}{
		{
			name: "Replication Controller Read Successfully",
		},
		{
			name:      "Read error",
			errString: "Failed to read from bundle",
			reader: fakeReaderWriter{
				readErr: fmt.Errorf("Failed to read from bundle"),
			},
		},
		{
			name:      "Write error",
			errString: "Failed to write bundle",
			writer: fakeReaderWriter{
				writeErr: fmt.Errorf("Failed to write bundle"),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			reader := NewManifestBundleReader(false, bundle, FactoryFromKubeConfigPath("testdata/kubeconfig.yaml"))
			if tt.reader != nil {
				reader.StreamReader.Reader = tt.reader
			}
			if tt.writer != nil {
				reader.writer = tt.writer
			}
			infos, err := reader.Read()
			if tt.errString != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
				require.Len(t, infos, 1)
				obj := infos[0].Object
				gvk := obj.GetObjectKind().GroupVersionKind()
				assert.Equal(t, gvk, schema.GroupVersionKind{
					Kind:    "ReplicationController",
					Group:   "",
					Version: "v1"})
			}
		})
	}
}

type fakeReaderWriter struct {
	readErr  error
	writeErr error
}

var _ io.Reader = fakeReaderWriter{}
var _ io.Writer = fakeReaderWriter{}

func (f fakeReaderWriter) Read(p []byte) (n int, err error) {
	return 0, f.readErr
}

func (f fakeReaderWriter) Write(p []byte) (n int, err error) {
	return 0, f.writeErr
}
