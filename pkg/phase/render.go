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
	"os"
	"strings"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/phase/errors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

const (
	// RenderSourceConfig will render a bundle that comes from site metadata file
	// and contains phase and executor docs
	RenderSourceConfig = "config"

	// RenderSourceExecutor indicates that rendering will be delegated to phase executor
	RenderSourceExecutor = "executor"

	// RenderSourcePhase the source will use kustomize root at phase entry point
	RenderSourcePhase = "phase"
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
	// Source identifies source of the bundle, these can be [phase|config|executor]
	// phase the source will use kustomize root at phase entry point
	// config will render a bundle that comes from site metadata file, and contains phase and executor docs
	// executor means that rendering will be delegated to phase executor
	Source string
	// FailOnDecryptionError makes sure that encrypted documents are getting decrypted by avoiding setting
	// env variable TOLERATE_DECRYPTION_FAILURES=true
	FailOnDecryptionError bool
	PhaseID               ifc.ID
}

// RunE prints out filtered documents
func (fo *RenderCommand) RunE(cfgFactory config.Factory, out io.Writer) error {
	if err := fo.Validate(); err != nil {
		return err
	}

	if !fo.FailOnDecryptionError {
		os.Setenv("TOLERATE_DECRYPTION_FAILURES", "true")
	}

	cfg, err := cfgFactory()
	if err != nil {
		return err
	}

	helper, err := NewHelper(cfg)
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

	if fo.Source == RenderSourceConfig {
		return renderConfigBundle(out, helper, sel)
	}

	client := NewClient(helper)
	phase, err := client.PhaseByID(fo.PhaseID)
	if err != nil {
		return err
	}

	var executorRender bool
	if fo.Source == RenderSourceExecutor {
		executorRender = true
	}
	return phase.Render(out, executorRender, ifc.RenderOptions{FilterSelector: sel})
}

func renderConfigBundle(out io.Writer, h ifc.Helper, sel document.Selector) error {
	bundle, err := document.NewBundleByPath(h.PhaseBundleRoot())
	if err != nil {
		return err
	}

	bundle, err = bundle.SelectBundle(sel)
	if err != nil {
		return err
	}
	return bundle.Write(out)
}

// Validate checks if command options are valid
func (fo *RenderCommand) Validate() (err error) {
	switch fo.Source {
	case RenderSourceConfig:
		// do nothing, source config doesnt need any parameters
	case RenderSourceExecutor, RenderSourcePhase:
		if fo.PhaseID.Name == "" {
			err = errors.ErrRenderPhaseNameNotSpecified{Sources: []string{RenderSourceExecutor, RenderSourcePhase}}
		}
	default:
		err = errors.ErrUnknownRenderSource{
			Source:       fo.Source,
			ValidSources: []string{RenderSourceConfig, RenderSourceExecutor, RenderSourcePhase},
		}
	}
	return err
}
