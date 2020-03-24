package utils

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func FactoryFromKubeConfigPath(kp string) cmdutil.Factory {
	kf := genericclioptions.NewConfigFlags(false)
	kf.KubeConfig = &kp
	return cmdutil.NewFactory(kf)
}
