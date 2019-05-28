package workflow_test

import (
	"bytes"
	"os"
	"testing"

	apixv1beta1fake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"

	"github.com/ian-howell/airshipctl/cmd"
	"github.com/ian-howell/airshipctl/cmd/workflow"
	argofake "github.com/ian-howell/airshipctl/pkg/client/clientset/versioned/fake"
	wfenv "github.com/ian-howell/airshipctl/pkg/workflow/environment"
	"github.com/ian-howell/airshipctl/test"
)

func TestWorkflowInit(t *testing.T) {
	actual := &bytes.Buffer{}
	rootCmd, settings, err := cmd.NewRootCmd(actual)
	if err != nil {
		t.Fatalf("Could not create root command: %s", err.Error())
	}
	workflowRoot := workflow.NewWorkflowCommand(actual, settings)
	workflowRoot.AddCommand(workflow.NewWorkflowInitCommand(actual, settings))
	argoClient := argofake.NewSimpleClientset()
	crdClient := apixv1beta1fake.NewSimpleClientset()
	kubeClient := kubefake.NewSimpleClientset()
	settings.PluginSettings[workflow.PluginSettingsID] = &wfenv.Settings{
		ArgoClient: argoClient,
		CRDClient:  crdClient,
		KubeClient: kubeClient,
	}
	rootCmd.AddCommand(workflowRoot)
	rootCmd.PersistentFlags().Parse(os.Args[1:])

	var tt test.CmdTest
	tt = test.CmdTest{
		Name:    "workflow-init",
		CmdLine: "workflow init",
	}

	test.RunTest(t, tt, rootCmd, actual)

	tt = test.CmdTest{
		Name:    "workflow-init-already-initialized",
		CmdLine: "workflow init",
	}

	test.RunTest(t, tt, rootCmd, actual)
}
