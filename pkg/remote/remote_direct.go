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

package remote

import (
	"context"

	api "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/phase"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

// DoRemoteDirect bootstraps the ephemeral node.
func (b baremetalHost) DoRemoteDirect(cfg *config.Config) error {
	helper, err := phase.NewHelper(cfg)
	if err != nil {
		return err
	}

	phaseClient := phase.NewClient(helper)
	phase, err := phaseClient.PhaseByID(ifc.ID{Name: config.BootstrapPhase})
	if err != nil {
		return err
	}

	docRoot, err := phase.DocumentRoot()
	if err != nil {
		return err
	}

	docBundle, err := document.NewBundleByPath(docRoot)
	if err != nil {
		return err
	}

	remoteDirectConfiguration := &api.RemoteDirectConfiguration{}
	selector, err := document.NewSelector().ByObject(remoteDirectConfiguration, api.Scheme)
	if err != nil {
		return err
	}
	doc, err := docBundle.SelectOne(selector)
	if err != nil {
		return err
	}

	err = doc.ToAPIObject(remoteDirectConfiguration, api.Scheme)
	if err != nil {
		return err
	}

	return b.RemoteDirect(context.Background(), remoteDirectConfiguration.IsoURL)
}
