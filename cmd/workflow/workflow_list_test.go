package workflow_test

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ian-howell/airshipctl/cmd"
	"github.com/ian-howell/airshipctl/cmd/workflow"
	"github.com/ian-howell/airshipctl/pkg/apis/workflow/v1alpha1"
	argofake "github.com/ian-howell/airshipctl/pkg/client/clientset/versioned/fake"
	"github.com/ian-howell/airshipctl/pkg/util"
	wfenv "github.com/ian-howell/airshipctl/pkg/workflow/environment"
	"github.com/ian-howell/airshipctl/test"
)

func TestWorkflowList(t *testing.T) {
	rootCmd, settings, err := cmd.NewRootCmd(nil)
	if err != nil {
		t.Fatalf("Could not create root command: %s", err.Error())
	}
	workflowRoot := workflow.NewWorkflowCommand(settings)
	workflowRoot.AddCommand(workflow.NewWorkflowListCommand(settings))
	rootCmd.AddCommand(workflowRoot)

	cmdTests := []WorkflowCmdTest{
		{
			CmdTest: &test.CmdTest{
				Name:    "workflow-list-empty",
				CmdLine: "workflow list",
				Objs:    []runtime.Object{},
			},
		},
		{
			CmdTest: &test.CmdTest{
				Name:    "workflow-list-nonempty",
				CmdLine: "workflow list",
				Objs:    []runtime.Object{},
			},
			ArgoObjs: []runtime.Object{
				&v1alpha1.Workflow{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fake-wf",
						CreationTimestamp: metav1.Time{
							Time: util.Clock().Add(5 * time.Minute),
						},
					},
					Status: v1alpha1.WorkflowStatus{
						Phase: "completed",
						StartedAt: metav1.Time{
							Time: util.Clock().Add(5 * time.Minute),
						},
						FinishedAt: metav1.Time{
							Time: util.Clock().Add(8 * time.Minute),
						},
					},
				},
			},
		},
	}

	for _, tt := range cmdTests {
		settings.PluginSettings[workflow.PluginSettingsID] = &wfenv.Settings{
			Initialized: true,
			ArgoClient:  argofake.NewSimpleClientset(tt.ArgoObjs...),
		}
		test.RunTest(t, tt.CmdTest, rootCmd)
	}
}
