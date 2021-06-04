/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package cluster

import (
	"fmt"
	"io"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/phase"
	"opendev.org/airship/airshipctl/pkg/util"
)

// StatusRunner runs internal logic of cluster status command
func StatusRunner(o StatusOptions, w io.Writer) error {
	statusMap, docs, err := o.GetStatusMapDocs()
	if err != nil {
		return err
	}

	var errors []error
	tw := util.NewTabWriter(w)
	fmt.Fprintf(tw, "Kind\tName\tStatus\n")
	for _, doc := range docs {
		status, err := statusMap.GetStatusForResource(doc)
		if err != nil {
			errors = append(errors, err)
		} else {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", doc.GetKind(), doc.GetName(), status)
		}
	}
	tw.Flush()

	if len(errors) > 0 {
		log.Debug("The following errors occurred while requesting the status:")
		for _, statusErr := range errors {
			log.Debug(statusErr)
		}
	}
	return nil
}

// GetKubeconfigCommand holds options for get kubeconfig command
type GetKubeconfigCommand struct {
	ClusterNames []string
	File         string
	Merge        bool
}

// RunE creates new kubeconfig interface object from secret, options hold the writer and merge(bool)
// to merge the kubeconfig. Writer in options is ignored if a file is provided.
func (cmd *GetKubeconfigCommand) RunE(cfgFactory config.Factory, writer io.Writer) error {
	cfg, err := cfgFactory()
	if err != nil {
		return err
	}

	helper, err := phase.NewHelper(cfg)
	if err != nil {
		return err
	}

	cMap, err := helper.ClusterMap()
	if err != nil {
		return err
	}

	var siteWide bool
	if len(cmd.ClusterNames) == 0 {
		siteWide = true
	}

	kubeconf := kubeconfig.NewBuilder().
		WithBundle(helper.PhaseConfigBundle()).
		WithClusterMap(cMap).
		WithClusterNames(cmd.ClusterNames...).
		WithTempRoot(helper.WorkDir()).
		SiteWide(siteWide).
		Build()

	if cmd.File != "" {
		return kubeconf.WriteFile(cmd.File, kubeconfig.WriteOptions{Merge: cmd.Merge})
	}
	return kubeconf.Write(writer)
}
