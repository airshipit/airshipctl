package workflow_test

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ian-howell/airshipctl/cmd"
	"github.com/ian-howell/airshipctl/cmd/workflow"
	"github.com/ian-howell/airshipctl/test"
)

type WorkflowCmdTest struct {
	*test.CmdTest

	CRDObjs  []runtime.Object
	ArgoObjs []runtime.Object
}

func TestWorkflow(t *testing.T) {
	rootCmd, settings, err := cmd.NewRootCmd(nil)
	if err != nil {
		t.Fatalf("Could not create root command: %s", err.Error())
	}
	workflowRoot := workflow.NewWorkflowCommand(settings)
	rootCmd.AddCommand(workflowRoot)

	cmdTests := []WorkflowCmdTest{
		{
			CmdTest: &test.CmdTest{
				Name:    "workflow",
				CmdLine: "workflow",
			},
		},
	}

	for _, tt := range cmdTests {
		test.RunTest(t, tt.CmdTest, rootCmd)
	}
}
