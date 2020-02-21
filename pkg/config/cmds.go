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
	"fmt"
	"io"
)

// RunGetAuthInfo performs the execution of 'config get-credentials' sub command
func RunGetAuthInfo(o *AuthInfoOptions, out io.Writer, airconfig *Config) error {
	if o.Name == "" {
		getAuthInfos(out, airconfig)
		return nil
	}
	return getAuthInfo(o, out, airconfig)
}

func getAuthInfo(o *AuthInfoOptions, out io.Writer, airconfig *Config) error {
	authinfo, err := airconfig.GetAuthInfo(o.Name)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, authinfo)
	return nil
}

func getAuthInfos(out io.Writer, airconfig *Config) {
	authinfos := airconfig.GetAuthInfos()
	if len(authinfos) == 0 {
		fmt.Fprintln(out, "No User credentials found in the configuration.")
	}
	for _, authinfo := range authinfos {
		fmt.Fprintln(out, authinfo)
	}
}

// RunGetCluster performs the execution of 'config get-cluster' sub command
func RunGetCluster(o *ClusterOptions, out io.Writer, airconfig *Config) error {
	if o.Name == "" {
		getClusters(out, airconfig)
		return nil
	}
	return getCluster(o.Name, o.ClusterType, out, airconfig)
}

func getCluster(cName, cType string, out io.Writer, airconfig *Config) error {
	cluster, err := airconfig.GetCluster(cName, cType)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s", cluster.PrettyString())
	return nil
}

func getClusters(out io.Writer, airconfig *Config) {
	clusters := airconfig.GetClusters()
	if len(clusters) == 0 {
		fmt.Fprintln(out, "No clusters found in the configuration.")
	}

	for _, cluster := range clusters {
		fmt.Fprintf(out, "%s\n", cluster.PrettyString())
	}
}

// RunGetContext performs the execution of 'config get-Context' sub command
func RunGetContext(o *ContextOptions, out io.Writer, airconfig *Config) error {
	if o.Name == "" && !o.CurrentContext {
		getContexts(out, airconfig)
		return nil
	}
	return getContext(o, out, airconfig)
}

func getContext(o *ContextOptions, out io.Writer, airconfig *Config) error {
	if o.CurrentContext {
		o.Name = airconfig.CurrentContext
	}
	context, err := airconfig.GetContext(o.Name)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s", context.PrettyString())
	return nil
}

func getContexts(out io.Writer, airconfig *Config) {
	contexts := airconfig.GetContexts()
	if len(contexts) == 0 {
		fmt.Fprintln(out, "No Contexts found in the configuration.")
	}
	for _, context := range contexts {
		fmt.Fprintf(out, "%s", context.PrettyString())
	}
}

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
