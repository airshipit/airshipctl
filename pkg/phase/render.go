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

package phase

import (
	"io"
	"strings"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

// RenderCommand holds filters for selector
type RenderCommand struct {
	// Label filters documents by label string
	Label string
	// Annotation filters documents by annotation string
	Annotation string
	// APIVersion filters documents by API group and version
	APIVersion string
	// Kind filters documents by document kind
	Kind string
	// Executor identifies if executor should perform rendering
	Executor bool
}

// RunE prints out filtered documents
func (fo *RenderCommand) RunE(cfgFactory config.Factory, phaseName string, out io.Writer) error {
	cfg, err := cfgFactory()
	if err != nil {
		return err
	}

	helper, err := NewHelper(cfg)
	if err != nil {
		return err
	}

	client := NewClient(helper)
	phase, err := client.PhaseByID(ifc.ID{Name: phaseName})
	if err != nil {
		return err
	}

	groupVersion := strings.Split(fo.APIVersion, "/")
	group := ""
	version := groupVersion[0]
	if len(groupVersion) > 1 {
		group = groupVersion[0]
		version = strings.Join(groupVersion[1:], "/")
	}

	sel := document.NewSelector().ByLabel(fo.Label).ByAnnotation(fo.Annotation).ByGvk(group, version, fo.Kind)
	return phase.Render(out, fo.Executor, ifc.RenderOptions{FilterSelector: sel})
}
