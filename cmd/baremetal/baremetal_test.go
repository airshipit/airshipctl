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

package baremetal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"opendev.org/airship/airshipctl/cmd/baremetal"
	"opendev.org/airship/airshipctl/testutil"
)

func TestBaremetal(t *testing.T) {
	tests := []*testutil.CmdTest{
		{
			Name:    "baremetal-with-help",
			CmdLine: "-h",
			Cmd:     baremetal.NewBaremetalCommand(nil),
		},
		{
			Name:    "baremetal-ejectmedia-with-help",
			CmdLine: "-h",
			Cmd:     baremetal.NewEjectMediaCommand(nil),
		},
		{
			Name:    "baremetal-poweroff-with-help",
			CmdLine: "-h",
			Cmd:     baremetal.NewPowerOffCommand(nil),
		},
		{
			Name:    "baremetal-poweron-with-help",
			CmdLine: "-h",
			Cmd:     baremetal.NewPowerOnCommand(nil),
		},
		{
			Name:    "baremetal-powerstatus-with-help",
			CmdLine: "-h",
			Cmd:     baremetal.NewPowerStatusCommand(nil),
		},
		{
			Name:    "baremetal-reboot-with-help",
			CmdLine: "-h",
			Cmd:     baremetal.NewRebootCommand(nil),
		},
		{
			Name:    "baremetal-remotedirect-with-help",
			CmdLine: "-h",
			Cmd:     baremetal.NewRemoteDirectCommand(nil),
		},
	}

	for _, tt := range tests {
		testutil.RunTest(t, tt)
	}
}

func TestGetHostSelectionsOneSelector(t *testing.T) {
	selectors := baremetal.GetHostSelections("node0", "")
	assert.Len(t, selectors, 1)
}

func TestGetHostSelectionsBothSelectors(t *testing.T) {
	selectors := baremetal.GetHostSelections("node0", "airshipit.org/ephemeral-node=true")
	assert.Len(t, selectors, 2)
}

func TestGetHostSelectionsNone(t *testing.T) {
	selectors := baremetal.GetHostSelections("", "")
	assert.Len(t, selectors, 0)
}
