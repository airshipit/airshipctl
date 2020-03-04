package initinfra

import (
	"sigs.k8s.io/kustomize/v3/pkg/fs"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/k8s/kubectl"
)

// Infra is an abstraction used to initialize base infrastructure
type Infra struct {
	FileSystem   fs.FileSystem
	RootSettings *environment.AirshipCTLSettings
	Client       client.Interface

	DryRun      bool
	Prune       bool
	ClusterType string
}

// NewInfra return instance of Infra
func NewInfra(rs *environment.AirshipCTLSettings) *Infra {
	// At this point AirshipCTLSettings may not be fully initialized
	infra := &Infra{RootSettings: rs}
	return infra
}

// Run intinfra subcommand logic
func (infra *Infra) Run() error {
	infra.FileSystem = kubectl.Buffer{FileSystem: fs.MakeRealFS()}
	var err error
	infra.Client, err = client.NewClient(infra.RootSettings)
	if err != nil {
		return err
	}
	return infra.Deploy()
}

// Deploy method deploys documents
func (infra *Infra) Deploy() error {
	kctl := infra.Client.Kubectl()
	var err error
	ao, err := kctl.ApplyOptions()
	if err != nil {
		return err
	}

	ao.SetDryRun(infra.DryRun)
	// If prune is true, set selector for purning
	if infra.Prune {
		ao.SetPrune(document.DeployedByLabel + "=" + document.InitinfraIdentifier)
	}

	globalConf := infra.RootSettings.Config()
	if err = globalConf.EnsureComplete(); err != nil {
		return err
	}

	var manifest *config.Manifest
	manifest, err = globalConf.CurrentContextManifest()
	if err != nil {
		return err
	}

	b, err := document.NewBundle(infra.FileSystem, manifest.TargetPath, "")
	if err != nil {
		return err
	}

	ls := document.EphemeralClusterSelector
	selector := document.NewSelector().ByLabel(ls)

	// Get documents that are annotated to belong to initinfra
	docs, err := b.Select(selector)
	if err != nil {
		return err
	}
	if len(docs) == 0 {
		return document.ErrDocNotFound{
			Selector: ls,
		}
	}

	// Label every document indicating that it was deployed by initinfra module for further reference
	// This may be used later to get all resources that are part of initinfra module, for monitoring, alerting
	// upgrading etc...
	// also if prune is set to true, this fulfills requirement for all labeled document to be labeled.
	// Pruning by annotation is not available, therefore we need to use label.
	for _, doc := range docs {
		res := doc.GetKustomizeResource()
		labels := res.GetLabels()
		labels[document.DeployedByLabel] = document.InitinfraIdentifier
		res.SetLabels(labels)
		err := doc.SetKustomizeResource(&res)
		if err != nil {
			return err
		}
	}

	return kctl.Apply(docs, ao)
}
