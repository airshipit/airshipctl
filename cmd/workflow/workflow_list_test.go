package workflow_test

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ian-howell/airshipctl/cmd"
	wfcmd "github.com/ian-howell/airshipctl/cmd/workflow"
	"github.com/ian-howell/airshipctl/pkg/apis/workflow/v1alpha1"
	"github.com/ian-howell/airshipctl/pkg/util"
	wf "github.com/ian-howell/airshipctl/pkg/workflow"
	wfenv "github.com/ian-howell/airshipctl/pkg/workflow/environment"
	"github.com/ian-howell/airshipctl/test"
)

func TestWorkflowList(t *testing.T) {
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
				Name:    "workflow-list-all-namespaces",
				CmdLine: "workflow list --all-namespaces",
				Objs:    []runtime.Object{},
			},
			ArgoObjs: []runtime.Object{
				&v1alpha1.Workflow{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fake-wf1",
						Namespace: "namespace1",
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
				&v1alpha1.Workflow{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fake-wf2",
						Namespace: "namespace2",
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
		{
			CmdTest: &test.CmdTest{
				Name:    "workflow-list-specific-namespace",
				CmdLine: "workflow list --namespace namespace1",
				Objs:    []runtime.Object{},
			},
			ArgoObjs: []runtime.Object{
				&v1alpha1.Workflow{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fake-wf1",
						Namespace: "namespace1",
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
				&v1alpha1.Workflow{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fake-wf2",
						Namespace: "namespace2",
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
		rootCmd, settings, err := cmd.NewRootCmd(nil)
		if err != nil {
			t.Fatalf("Could not create root command: %s", err.Error())
		}
		workflowRoot := wfcmd.NewWorkflowCommand(settings)
		workflowRoot.AddCommand(wfcmd.NewWorkflowListCommand(&wfenv.Settings{AirshipCTLSettings: settings}))
		rootCmd.AddCommand(workflowRoot)

		// This will initialize the singleton clientset as a mock
		wf.NewSimpleClientset(tt.Objs, tt.ArgoObjs, tt.CRDObjs)
		test.RunTest(t, tt.CmdTest, rootCmd)
	}
}
