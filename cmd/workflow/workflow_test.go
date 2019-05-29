package workflow_test

import (
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ian-howell/airshipctl/test"
)

type WorkflowCmdTest struct {
	*test.CmdTest

	CRDObjs  []runtime.Object
	ArgoObjs []runtime.Object
}
