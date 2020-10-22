// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package main implements an injection function for resource reservations and
// is run with `kustomize config run -- DIR/`.
package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"

	"opendev.org/airship/airshipctl/pkg/document/plugin/templater"
)

func main() {
	cfg := make(map[string]interface{})
	resourceList := &framework.ResourceList{FunctionConfig: &cfg}
	cmd := framework.Command(resourceList, func() error {
		plugin, err := templater.New(cfg)
		if err != nil {
			return err
		}
		resourceList.Items, err = plugin.Filter(resourceList.Items)
		return err
	})
	if err := cmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
