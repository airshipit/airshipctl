/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"fmt"
	"io"
)

// ContextOption is a function that allows to modify context object
type ContextOption func(ctx *Context)

// SetContextManifest sets manifest in context object
func SetContextManifest(manifest string) ContextOption {
	return func(ctx *Context) {
		ctx.Manifest = manifest
	}
}

// SetContextManagementConfig sets management config in context object
func SetContextManagementConfig(managementConfig string) ContextOption {
	return func(ctx *Context) {
		ctx.ManagementConfiguration = managementConfig
	}
}

// RunSetContextOptions are options required to create/modify airshipctl context
type RunSetContextOptions struct {
	CfgFactory Factory
	CtxName    string
	Current    bool
	Writer     io.Writer
}

// RunSetContext validates the given command line options and invokes AddContext/ModifyContext
func (o *RunSetContextOptions) RunSetContext(opts ...ContextOption) error {
	cfg, err := o.CfgFactory()
	if err != nil {
		return err
	}

	if o.Current {
		o.CtxName = cfg.CurrentContext
	}

	if o.CtxName == "" {
		return ErrEmptyContextName{}
	}

	infoMsg := fmt.Sprintf("context with name %s", o.CtxName)
	context, err := cfg.GetContext(o.CtxName)
	if err != nil {
		// context didn't exist, create it
		cfg.AddContext(o.CtxName, opts...)
		infoMsg = fmt.Sprintf("%s created\n", infoMsg)
	} else {
		// Context exists, lets update
		cfg.ModifyContext(context, opts...)
		infoMsg = fmt.Sprintf("%s modified\n", infoMsg)
	}

	// Verify we didn't break anything
	if err = cfg.EnsureComplete(); err != nil {
		return err
	}

	if _, err := o.Writer.Write([]byte(infoMsg)); err != nil {
		return err
	}
	// Update configuration file just in time persistence approach
	return cfg.PersistConfig(true)
}

// RunUseContext validates the given context name and updates it as current context
func RunUseContext(desiredContext string, airconfig *Config) error {
	if _, err := airconfig.GetContext(desiredContext); err != nil {
		return err
	}

	if airconfig.CurrentContext != desiredContext {
		airconfig.CurrentContext = desiredContext
		if err := airconfig.PersistConfig(true); err != nil {
			return err
		}
	}
	return nil
}

// RunSetManifest validates the given command line options and invokes AddManifest/ModifyManifest
func RunSetManifest(o *ManifestOptions, airconfig *Config, writeToStorage bool) (bool, error) {
	modified := false
	err := o.Validate()
	if err != nil {
		return modified, err
	}

	manifest, exists := airconfig.Manifests[o.Name]
	if !exists {
		// manifest didn't exist, create it
		// ignoring the returned added manifest
		airconfig.AddManifest(o)
	} else {
		// manifest exists, lets update
		err = airconfig.ModifyManifest(manifest, o)
		if err != nil {
			return modified, err
		}
		modified = true
	}
	// Update configuration file just in time persistence approach
	if writeToStorage {
		if err := airconfig.PersistConfig(true); err != nil {
			// Error that it didnt persist the changes
			return modified, ErrConfigFailed{}
		}
	}

	return modified, nil
}
