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

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/util"
	"opendev.org/airship/airshipctl/pkg/version"
)

// NewVersionCommand creates a command for displaying the version of airshipctl
func NewVersionCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Airshipctl command to display the current version number",
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			clientV := clientVersion()
			w := util.GetNewTabWriter(out)
			defer w.Flush()
			fmt.Fprintf(w, "%s:\t%s\n", "airshipctl", clientV)
		},
	}
	return versionCmd
}

func clientVersion() string {
	v := version.Get()
	return fmt.Sprintf("%#v", v)
}
