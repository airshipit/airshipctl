// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package main implements an injection function for resource reservations and
// is run with `kustomize config run -- DIR/`.
package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"

	"opendev.org/airship/airshipctl/pkg/document/plugin/templater"
)

func main() {
	fn := func(rl *framework.ResourceList) error {
		cfg, err := rl.FunctionConfig.Map()
		if err != nil {
			return err
		}
		plugin, err := templater.New(cfg)
		if err != nil {
			return err
		}
		rl.Items, err = plugin.Filter(rl.Items)
		return err
	}
	cmd := command.Build(framework.ResourceListProcessorFunc(fn), command.StandaloneEnabled, false)
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
