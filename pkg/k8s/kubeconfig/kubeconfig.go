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

package kubeconfig

// File determines where kubeconfig is located on file system and which context to use
type File struct {
	Path    string
	Context string
}

// Provider interface allows to get kubeconfig file path and context based on cluster type
type Provider interface {
	// If clusterType is an empty string it means that caller is not aware then default cluster type will be used
	// default cluster type maybe different for different provider implementations, for example if we are providing
	// kubeconfig file for a phase then phase may be bound to ephemeral or target cluster type then defaults will be
	// ephemeral or target respectively.
	Get(clusterType string) (File, error)
	Cleanup() error
}
