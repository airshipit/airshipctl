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

package plan

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/errors"
)

const (
	runLong = `
Run life-cycle phase plan which was defined in document model.
`
)

// NewRunCommand creates a command which execute a particular phase plan
func NewRunCommand(cfgFactory config.Factory) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "run PLAN_NAME",
		Short: "Run plan",
		Long:  runLong[1:],
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.ErrNotImplemented{What: "airshipctl plan run"}
		},
	}
	return listCmd
}
