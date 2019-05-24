package workflow_test

import (
	"bytes"
	"os"
	"testing"

	argofake "github.com/ian-howell/airshipctl/pkg/workflow/clientset/versioned/fake"

	"github.com/ian-howell/airshipctl/cmd"
	"github.com/ian-howell/airshipctl/cmd/workflow"
	"github.com/ian-howell/airshipctl/test"
	"github.com/ian-howell/airshipctl/pkg/environment"
	wfenv "github.com/ian-howell/airshipctl/pkg/workflow/environment"
)

func TestWorkflowList(t *testing.T) {
	tests := []test.CmdTest{
		{
			Name:    "workflow-list",
			CmdLine: "workflow list",
		},
	}
	for _, tt := range tests {
		actual := &bytes.Buffer{}
		rootCmd, err := cmd.NewRootCmd(actual)
		if err != nil {
			t.Fatalf("Could not create root command: %s", err.Error())
		}
		settings := &environment.AirshipCTLSettings{}
		settings.InitFlags(rootCmd)
		workflowRoot := workflow.NewWorkflowCommand(actual, settings)
		workflowRoot.AddCommand(workflow.NewWorkflowListCommand(actual, settings))
		settings.PluginSettings[workflow.PluginSettingsID] = &wfenv.Settings{
			ArgoClient: argofake.NewSimpleClientset(),
		}

		rootCmd.AddCommand(workflowRoot)
		rootCmd.PersistentFlags().Parse(os.Args[1:])
		test.RunTest(t, tt, rootCmd, actual)
	}
}
