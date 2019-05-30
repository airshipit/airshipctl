package workflow_test

import (
	"testing"

	v1beta2 "k8s.io/api/apps/v1beta2"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apixv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apixv1beta1fake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"

	"github.com/ian-howell/airshipctl/cmd"
	"github.com/ian-howell/airshipctl/cmd/workflow"
	argofake "github.com/ian-howell/airshipctl/pkg/client/clientset/versioned/fake"
	wfenv "github.com/ian-howell/airshipctl/pkg/workflow/environment"
	"github.com/ian-howell/airshipctl/test"
)

func TestWorkflowInit(t *testing.T) {
	rootCmd, settings, err := cmd.NewRootCmd(nil)
	if err != nil {
		t.Fatalf("Could not create root command: %s", err.Error())
	}
	workflowRoot := workflow.NewWorkflowCommand(settings)
	workflowRoot.AddCommand(workflow.NewWorkflowInitCommand(settings))
	rootCmd.AddCommand(workflowRoot)

	cmdTests := []WorkflowCmdTest{
		{
			CmdTest: &test.CmdTest{
				Name:    "workflow-init",
				CmdLine: "workflow init",
				Objs:    []runtime.Object{},
			},
		},
		{
			CmdTest: &test.CmdTest{
				Name:    "workflow-init-already-initialized",
				CmdLine: "workflow init",
				Objs: []runtime.Object{
					&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "argo"}},
					&v1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "argo", Namespace: "argo"}},
					&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "argo-aggregate-to-admin"}},
					&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "argo-aggregate-to-edit"}},
					&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "argo-aggregate-to-view"}},
					&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "argo-cluster-role"}},
					&rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "argo-binding"}},
					&v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "workflow-controller-configmap", Namespace: "argo"}},
					&v1beta2.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "workflow-controller", Namespace: "argo"}},
				},
			},
			ArgoObjs: []runtime.Object{},
			CRDObjs: []runtime.Object{
				&apixv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "workflows.argoproj.io"}},
			},
		},
	}

	for _, tt := range cmdTests {
		settings.PluginSettings[workflow.PluginSettingsID] = &wfenv.Settings{
			Initialized: true,
			KubeClient:  kubefake.NewSimpleClientset(tt.CmdTest.Objs...),
			ArgoClient:  argofake.NewSimpleClientset(tt.ArgoObjs...),
			CRDClient:   apixv1beta1fake.NewSimpleClientset(tt.CRDObjs...),
		}
		test.RunTest(t, tt.CmdTest, rootCmd)
	}
}
