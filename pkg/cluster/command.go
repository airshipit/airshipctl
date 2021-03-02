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

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/clusterctl/client"
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

// GetKubeconfig creates new kubeconfig interface object from secret and prints its content to writer
func GetKubeconfig(cfgFactory config.Factory, clusterName string, writer io.Writer) error {
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

	wd, err := helper.WorkDir()
	if err != nil {
		return err
	}

	client, err := client.NewClient(helper.TargetPath(), log.DebugEnabled(), airshipv1.DefaultClusterctl())
	if err != nil {
		return err
	}

	kubeconf := kubeconfig.NewBuilder().
		WithBundle(helper.PhaseBundleRoot()).
		WithClusterctClient(client).
		WithClusterMap(cMap).
		WithClusterName(clusterName).
		WithTempRoot(wd).
		Build()

	return kubeconf.Write(writer)
}
