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

package document

// Label Selectors
const (
	BaseAirshipSelector       = "airshipit.org"
	EphemeralHostSelector     = BaseAirshipSelector + "/ephemeral-node in (True, true)"
	EphemeralUserDataSelector = BaseAirshipSelector + "/ephemeral-user-data in (True, true)"
	InitInfraSelector         = BaseAirshipSelector + "/phase = initinfra"

	// DeployToK8sSelector checks that deploy-k8s label is not equal to true or True (string)
	// Please note that by default every document in the manifest is to be deployed to kubernetes cluster.
	DeployToK8sSelector = "config.kubernetes.io/local-config notin (True, true)"
)

// Kinds
const (
	SecretKind        = "Secret"
	BareMetalHostKind = "BareMetalHost"
)
