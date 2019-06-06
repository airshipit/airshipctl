package workflow_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ian-howell/airshipctl/pkg/apis/workflow/v1alpha1"
	wf "github.com/ian-howell/airshipctl/pkg/workflow"
)

const testNamespace = "testNamespace"

func TestListWorkflows(t *testing.T) {
	tests := []struct {
		Name     string
		KubeObjs []runtime.Object
		CRDObjs  []runtime.Object
		ArgoObjs []runtime.Object
	}{
		{
			Name: "no workflows",
		},
		{
			Name: "One workflow",
			ArgoObjs: []runtime.Object{
				&v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Name: "testWorkflow", Namespace: testNamespace}},
			},
		},
		{
			Name: "Multiple workflow",
			ArgoObjs: []runtime.Object{
				&v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Name: "testWorkflow1", Namespace: testNamespace}},
				&v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Name: "testWorkflow2", Namespace: testNamespace}},
				&v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Name: "testWorkflow3", Namespace: testNamespace}},
				&v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Name: "testWorkflow4", Namespace: testNamespace}},
				&v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Name: "testWorkflow5", Namespace: testNamespace}},
			},
		},
	}
	for _, test := range tests {
		clientset := wf.NewSimpleClientset(test.KubeObjs, test.ArgoObjs, test.CRDObjs)
		wflist, err := clientset.Argo.ArgoprojV1alpha1().Workflows(testNamespace).List(v1.ListOptions{})
		if err != nil {
			t.Fatalf("Unexpected error: %s", err.Error())
		}
		if len(wflist.Items) != len(test.ArgoObjs) {
			t.Errorf("Expected %d workflows, got %d", len(test.ArgoObjs), len(wflist.Items))
		}
		for _, expected := range test.ArgoObjs {
			found := false
			expectedName := expected.(*v1alpha1.Workflow).Name
			for _, actual := range wflist.Items {
				if actual.Name == expectedName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Missing Workflow '%s'", expectedName)
			}
		}
	}
}
