/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugins_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/testutil"
)

func TestLoaderConfig(t *testing.T) {
	t.Run("Try load non-existent plugin", func(t *testing.T) {
		_, err := document.NewBundle(testutil.SetupTestFs(t, "testdata/unknownplugin"), "/", "/")
		assert.Error(t, err)
	})
}
