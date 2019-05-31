package util

import (
	"github.com/ian-howell/airshipctl/pkg/apis/workflow/v1alpha1"
)

// IsWorkflowSuspended returns whether or not a workflow is considered suspended
func IsWorkflowSuspended(wf *v1alpha1.Workflow) bool {
	if wf.Spec.Suspend != nil && *wf.Spec.Suspend {
		return true
	}
	for _, node := range wf.Status.Nodes {
		if node.Type == v1alpha1.NodeTypeSuspend && node.Phase == v1alpha1.NodeRunning {
			return true
		}
	}
	return false
}

// IsWorkflowTerminated returns whether or not a workflow is considered terminated
func IsWorkflowTerminated(wf *v1alpha1.Workflow) bool {
	return wf.Spec.ActiveDeadlineSeconds != nil && *wf.Spec.ActiveDeadlineSeconds == 0
}
