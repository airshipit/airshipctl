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
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/log"
)

// DoRemoteDirect bootstraps the ephemeral node.
func (b baremetalHost) DoRemoteDirect(settings *environment.AirshipCTLSettings) error {
	cfg := settings.Config
	bootstrapSettings, err := cfg.CurrentContextBootstrapInfo()
	if err != nil {
		return err
	}

	remoteConfig := bootstrapSettings.RemoteDirect
	if remoteConfig == nil {
		return config.ErrMissingConfig{What: "RemoteDirect options not defined in bootstrap config"}
	}

	log.Debugf("Using ephemeral node %s with BMC Address %s", b.NodeID(), b.BMCAddress)

	// Perform remote direct operations
	if remoteConfig.IsoURL == "" {
		return ErrMissingBootstrapInfoOption{What: "isoURL"}
	}

	err = b.SetVirtualMedia(b.Context, remoteConfig.IsoURL)
	if err != nil {
		return err
	}

	log.Debugf("Successfully loaded virtual media: %q", remoteConfig.IsoURL)

	err = b.SetBootSourceByType(b.Context)
	if err != nil {
		return err
	}

	err = b.RebootSystem(b.Context)
	if err != nil {
		return err
	}

	log.Debugf("Restarted ephemeral host %s", b.NodeID())

	return nil
}
