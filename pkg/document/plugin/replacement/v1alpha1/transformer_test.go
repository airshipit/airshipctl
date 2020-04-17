// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	replv1alpha1 "opendev.org/airship/airshipctl/pkg/document/plugin/replacement/v1alpha1"
	plugtypes "opendev.org/airship/airshipctl/pkg/document/plugin/types"
)

func samplePlugin(t *testing.T) plugtypes.Plugin {
	plugin, err := replv1alpha1.New(nil, []byte(`
apiVersion: airshipit.org/v1beta1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    value: nginx:newtag
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers[name=nginx-latest].image`))

	require.NoError(t, err)
	return plugin
}

func TestMalformedConfig(t *testing.T) {
	_, err := replv1alpha1.New(nil, []byte("--"))
	assert.Error(t, err)
}

func TestMalformedInput(t *testing.T) {
	plugin := samplePlugin(t)
	err := plugin.Run(strings.NewReader("--"), &bytes.Buffer{})
	assert.Error(t, err)
}

func TestDuplicatedResources(t *testing.T) {
	plugin := samplePlugin(t)
	err := plugin.Run(strings.NewReader(`
apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - name: myapp-container
    image: busybox
---
apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - name: myapp-container
    image: busybox
 `), &bytes.Buffer{})
	assert.Error(t, err)
}

func TestReplacementTransformer(t *testing.T) {
	testCases := []struct {
		cfg         string
		in          string
		expectedOut string
	}{
		{
			cfg: `
apiVersion: airshipit.org/v1beta1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    value: nginx:newtag
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers[name=nginx-latest].image
- source:
    value: postgres:latest
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers.3.image
`,

			in: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:latest
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`,
			expectedOut: `apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:newtag
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:latest
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1beta1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    objref:
      kind: Pod
      name: pod
    fieldref: spec.containers
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers`,
			in: `
apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - name: myapp-container
    image: busybox
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy3
`,
			expectedOut: `apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - image: busybox
    name: myapp-container
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: busybox
        name: myapp-container
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy3
spec:
  template:
    spec:
      containers:
      - image: busybox
        name: myapp-container
`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1beta1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    objref:
      kind: ConfigMap
      name: cm
    fieldref: data.HOSTNAME
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers[image=debian].args.0
    - spec.template.spec.containers[name=busybox].args.1
- source:
    objref:
      kind: ConfigMap
      name: cm
    fieldref: data.PORT
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers[image=debian].args.1
    - spec.template.spec.containers[name=busybox].args.2`,
			in: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
  labels:
    foo: bar
spec:
  template:
    metadata:
      labels:
        foo: bar
    spec:
      containers:
        - name: command-demo-container
          image: debian
          command: ["printenv"]
          args:
            - HOSTNAME
            - PORT
        - name: busybox
          image: busybox:latest
          args:
            - echo
            - HOSTNAME
            - PORT
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOSTNAME: example.com
  PORT: 8080`,
			expectedOut: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    foo: bar
  name: deploy
spec:
  template:
    metadata:
      labels:
        foo: bar
    spec:
      containers:
      - args:
        - example.com
        - 8080
        command:
        - printenv
        image: debian
        name: command-demo-container
      - args:
        - echo
        - example.com
        - 8080
        image: busybox:latest
        name: busybox
---
apiVersion: v1
data:
  HOSTNAME: example.com
  PORT: 8080
kind: ConfigMap
metadata:
  name: cm
`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1beta1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    value: regexedtag
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers[name=nginx-latest].image%TAG%
- source:
    value: postgres:latest
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers.3.image`,
			in: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:TAG
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine`,
			expectedOut: `apiVersion: v1
group: apps
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:regexedtag
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:latest
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`,
		},
	}

	for _, tc := range testCases {
		plugun, err := replv1alpha1.New(nil, []byte(tc.cfg))
		require.NoError(t, err)

		buf := &bytes.Buffer{}
		err = plugun.Run(strings.NewReader(tc.in), buf)
		require.NoError(t, err)
		assert.Equal(t, tc.expectedOut, buf.String())
	}
}
