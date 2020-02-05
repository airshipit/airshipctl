/*l
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

package config

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/errors"
)

var (
	configInitLong = "Generate initial configuration files for airshipctl"
)

// NewCmdConfigInit returns a Command instance for 'config init' sub command
func NewCmdConfigInit(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	configInitCmd := &cobra.Command{
		Use:   "init",
		Short: configInitLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.ErrNotImplemented{}
		},
	}

	return configInitCmd
}
