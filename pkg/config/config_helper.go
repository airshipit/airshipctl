/*
Copyright 2014 The Kubernetes Authors.

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
	"errors"
)

// RunSetAuthInfo validates the given command line options and invokes AddAuthInfo/ModifyAuthInfo
func RunSetAuthInfo(o *AuthInfoOptions, airconfig *Config, writeToStorage bool) (bool, error) {
	modified := false
	err := o.Validate()
	if err != nil {
		return modified, err
	}

	authinfo, err := airconfig.GetAuthInfo(o.Name)
	if err != nil {
		var cerr ErrMissingConfig
		if !errors.As(err, &cerr) {
			// An error occurred, but it wasn't a "missing" config error.
			return modified, err
		}

		// authinfo didn't exist, create it
		// ignoring the returned added authinfo
		airconfig.AddAuthInfo(o)
	} else {
		// AuthInfo exists, lets update
		airconfig.ModifyAuthInfo(authinfo, o)
		modified = true
	}
	// Update configuration file just in time persistence approach
	if writeToStorage {
		if err := airconfig.PersistConfig(); err != nil {
			// Error that it didnt persist the changes
			return modified, ErrConfigFailed{}
		}
	}

	return modified, nil
}

// RunSetCluster validates the given command line options and invokes AddCluster/ModifyCluster
func RunSetCluster(o *ClusterOptions, airconfig *Config, writeToStorage bool) (bool, error) {
	modified := false
	err := o.Validate()
	if err != nil {
		return modified, err
	}

	cluster, err := airconfig.GetCluster(o.Name, o.ClusterType)
	if err != nil {
		var cerr ErrMissingConfig
		if !errors.As(err, &cerr) {
			// An error occurred, but it wasn't a "missing" config error.
			return modified, err
		}

		// Cluster didn't exist, create it
		_, err := airconfig.AddCluster(o)
		if err != nil {
			return modified, err
		}
		modified = false
	} else {
		// Cluster exists, lets update
		_, err := airconfig.ModifyCluster(cluster, o)
		if err != nil {
			return modified, err
		}
		modified = true
	}

	// Update configuration file
	// Just in time persistence approach
	if writeToStorage {
		if err := airconfig.PersistConfig(); err != nil {
			// Some warning here , that it didnt persist the changes because of this
			// Or should we float this up
			// What would it mean? No value.
			return modified, err
		}
	}

	return modified, nil
}

// RunSetContext validates the given command line options and invokes AddContext/ModifyContext
func RunSetContext(o *ContextOptions, airconfig *Config, writeToStorage bool) (bool, error) {
	modified := false
	err := o.Validate()
	if err != nil {
		return modified, err
	}
	if o.Current {
		if airconfig.CurrentContext == "" {
			return modified, ErrMissingCurrentContext{}
		}
		// when --current flag is passed, use current context
		o.Name = airconfig.CurrentContext
	}

	context, err := airconfig.GetContext(o.Name)
	if err != nil {
		var cerr ErrMissingConfig
		if !errors.As(err, &cerr) {
			// An error occurred, but it wasn't a "missing" config error.
			return modified, err
		}

		if o.CurrentContext {
			return modified, ErrMissingConfig{}
		}
		// context didn't exist, create it
		// ignoring the returned added context
		airconfig.AddContext(o)
	} else {
		// Found the desired Current Context
		// Lets update it and be done.
		if o.CurrentContext {
			airconfig.CurrentContext = o.Name
		} else {
			// Context exists, lets update
			airconfig.ModifyContext(context, o)
		}
		modified = true
	}
	// Update configuration file just in time persistence approach
	if writeToStorage {
		if err := airconfig.PersistConfig(); err != nil {
			// Error that it didnt persist the changes
			return modified, ErrConfigFailed{}
		}
	}

	return modified, nil
}

// RunUseContext validates the given context name and updates it as current context
func RunUseContext(desiredContext string, airconfig *Config) error {
	if _, err := airconfig.GetContext(desiredContext); err != nil {
		return err
	}

	if airconfig.CurrentContext != desiredContext {
		airconfig.CurrentContext = desiredContext
		if err := airconfig.PersistConfig(); err != nil {
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
		if err := airconfig.PersistConfig(); err != nil {
			// Error that it didnt persist the changes
			return modified, ErrConfigFailed{}
		}
	}

	return modified, nil
}
