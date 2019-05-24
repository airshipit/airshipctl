package workflow_test

import (
	"bytes"
	"os"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ian-howell/airshipctl/cmd"
	"github.com/ian-howell/airshipctl/cmd/workflow"
	"github.com/ian-howell/airshipctl/pkg/workflow/apis/workflow/v1alpha1"
	argofake "github.com/ian-howell/airshipctl/pkg/workflow/clientset/versioned/fake"
	wfenv "github.com/ian-howell/airshipctl/pkg/workflow/environment"
	"github.com/ian-howell/airshipctl/test"
)

func TestWorkflowList(t *testing.T) {
	actual := &bytes.Buffer{}
	rootCmd, settings, err := cmd.NewRootCmd(actual)
	if err != nil {
		t.Fatalf("Could not create root command: %s", err.Error())
	}
	workflowRoot := workflow.NewWorkflowCommand(actual, settings)
	workflowRoot.AddCommand(workflow.NewWorkflowListCommand(actual, settings))
	argoClient := argofake.NewSimpleClientset()
	settings.PluginSettings[workflow.PluginSettingsID] = &wfenv.Settings{
		ArgoClient: argoClient,
	}
	rootCmd.AddCommand(workflowRoot)
	rootCmd.PersistentFlags().Parse(os.Args[1:])

	var tt test.CmdTest
	tt = test.CmdTest{
		Name:    "workflow-list-empty",
		CmdLine: "workflow list",
	}

	test.RunTest(t, tt, rootCmd, actual)

	argoClient.ArgoprojV1alpha1().Workflows("default").Create(&v1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake-wf",
		},
		Status: v1alpha1.WorkflowStatus{
			Phase: "completed",
		},
	})

	tt = test.CmdTest{
		Name:    "workflow-list-nonempty",
		CmdLine: "workflow list",
	}
	test.RunTest(t, tt, rootCmd, actual)
}
