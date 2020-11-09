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

package ifc

// BaremetalHostSelector allows to select baremetal hosts, if used empty all possible hosts
// will be should be returned by Select() method of BaremetalInvenotry interface
type BaremetalHostSelector struct {
	PhaseSelector PhaseSelector
	Name          string
	Namespace     string
	LabelSelector string
}

// PhaseSelector allows to select hosts based on phase they belong to
type PhaseSelector struct {
	Name      string
	Namespace string
}

// ByName allows to select hosts based on their name
func (s BaremetalHostSelector) ByName(name string) BaremetalHostSelector {
	s.Name = name
	return s
}

// ByNamespace allows to select hosts based on their namespace
func (s BaremetalHostSelector) ByNamespace(name, namespace string) BaremetalHostSelector {
	s.PhaseSelector = PhaseSelector{
		Name:      name,
		Namespace: namespace,
	}
	return s
}

// ByLabel allows to select hosts based on label
func (s BaremetalHostSelector) ByLabel(label string) BaremetalHostSelector {
	s.LabelSelector = label
	return s
}
