package workflow

import (
	"sort"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ian-howell/airshipctl/pkg/apis/workflow/v1alpha1"
	v1alpha1client "github.com/ian-howell/airshipctl/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/ian-howell/airshipctl/pkg/workflow/environment"
)

// ListWorkflows returns a list of Workflows
func ListWorkflows(settings *environment.Settings) ([]v1alpha1.Workflow, error) {
	var clientSet v1alpha1client.WorkflowInterface
	if settings.AllNamespaces {
		clientSet = settings.ArgoClient.ArgoprojV1alpha1().Workflows(apiv1.NamespaceAll)
	} else {
		clientSet = settings.ArgoClient.ArgoprojV1alpha1().Workflows(settings.Namespace)
	}
	wflist, err := clientSet.List(v1.ListOptions{})
	if err != nil {
		return []v1alpha1.Workflow{}, err
	}
	workflows := wflist.Items
	sort.Sort(ByFinishedAt(workflows))
	return workflows, nil
}

// ByFinishedAt is a sort interface which sorts running jobs earlier before considering FinishedAt
type ByFinishedAt []v1alpha1.Workflow

func (f ByFinishedAt) Len() int      { return len(f) }
func (f ByFinishedAt) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f ByFinishedAt) Less(i, j int) bool {
	iStart := f[i].ObjectMeta.CreationTimestamp
	iFinish := f[i].Status.FinishedAt
	jStart := f[j].ObjectMeta.CreationTimestamp
	jFinish := f[j].Status.FinishedAt
	if iFinish.IsZero() && jFinish.IsZero() {
		return !iStart.Before(&jStart)
	}
	if iFinish.IsZero() && !jFinish.IsZero() {
		return true
	}
	if !iFinish.IsZero() && jFinish.IsZero() {
		return false
	}
	return jFinish.Before(&iFinish)
}
