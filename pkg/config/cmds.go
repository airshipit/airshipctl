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
	"io/ioutil"
)

// Validate that the arguments are correct
func (o *ClusterOptions) Validate() error {
	if len(o.Name) == 0 {
		return errors.New("you must specify a non-empty cluster name")
	}
	err := ValidClusterType(o.ClusterType)
	if err != nil {
		return err
	}
	if o.InsecureSkipTLSVerify && o.CertificateAuthority != "" {
		return fmt.Errorf("you cannot specify a %s and %s mode at the same time", FlagCAFile, FlagInsecure)
	}

	if !o.EmbedCAData {
		return nil
	}
	caPath := o.CertificateAuthority
	if caPath == "" {
		return fmt.Errorf("you must specify a --%s to embed", FlagCAFile)
	}
	if _, err := ioutil.ReadFile(caPath); err != nil {
		return fmt.Errorf("could not read %s data from %s: %v", FlagCAFile, caPath, err)
	}
	return nil
}

func (o *ContextOptions) Validate() error {
	if len(o.Name) == 0 {
		return errors.New("you must specify a non-empty context name")
	}
	// Expect ClusterType only when this is not setting currentContext
	if o.ClusterType != "" {
		err := ValidClusterType(o.ClusterType)
		if err != nil {
			return err
		}
	}
	// TODO Manifest, Cluster could be validated against the existing config maps
	return nil
}

func (o *AuthInfoOptions) Validate() error {
	if len(o.Token) > 0 && (len(o.Username) > 0 || len(o.Password) > 0) {
		return fmt.Errorf("you cannot specify more than one authentication method at the same time: --%v  or --%v/--%v",
			FlagBearerToken, FlagUsername, FlagPassword)
	}
	if !o.EmbedCertData {
		return nil
	}
	certPath := o.ClientCertificate
	if certPath == "" {
		return fmt.Errorf("you must specify a --%s to embed", FlagCertFile)
	}
	if _, err := ioutil.ReadFile(certPath); err != nil {
		return fmt.Errorf("error reading %s data from %s: %v", FlagCertFile, certPath, err)
	}
	keyPath := o.ClientKey
	if keyPath == "" {
		return fmt.Errorf("you must specify a --%s to embed", FlagKeyFile)
	}
	if _, err := ioutil.ReadFile(keyPath); err != nil {
		return fmt.Errorf("error reading %s data from %s: %v", FlagKeyFile, keyPath, err)
	}
	return nil
}

// runGetAuthInfo performs the execution of 'config get-credentials' sub command
func RunGetAuthInfo(o *AuthInfoOptions, out io.Writer, airconfig *Config) error {
	if o.Name == "" {
		return getAuthInfos(out, airconfig)
	}
	return getAuthInfo(o, out, airconfig)
}

func getAuthInfo(o *AuthInfoOptions, out io.Writer, airconfig *Config) error {
	cName := o.Name
	authinfo, err := airconfig.GetAuthInfo(cName)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, authinfo)
	return nil
}

func getAuthInfos(out io.Writer, airconfig *Config) error {
	authinfos, err := airconfig.GetAuthInfos()
	if err != nil {
		return err
	}
	if len(authinfos) == 0 {
		fmt.Fprintln(out, "No User credentials found in the configuration.")
	}
	for _, authinfo := range authinfos {
		fmt.Fprintln(out, authinfo)
	}
	return nil
}

// runGetCluster performs the execution of 'config get-cluster' sub command
func RunGetCluster(o *ClusterOptions, out io.Writer, airconfig *Config) error {
	if o.Name == "" {
		return getClusters(out, airconfig)
	}
	return getCluster(o.Name, o.ClusterType, out, airconfig)
}

func getCluster(cName, cType string,
	out io.Writer, airconfig *Config) error {
	cluster, err := airconfig.GetCluster(cName, cType)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s", cluster.PrettyString())
	return nil
}

func getClusters(out io.Writer, airconfig *Config) error {
	clusters, err := airconfig.GetClusters()
	if err != nil {
		return err
	}
	if len(clusters) == 0 {
		fmt.Fprintln(out, "No clusters found in the configuration.")
		return nil
	}

	for _, cluster := range clusters {
		fmt.Fprintf(out, "%s\n", cluster.PrettyString())
	}
	return nil
}

// runGetContext performs the execution of 'config get-Context' sub command
func RunGetContext(o *ContextOptions, out io.Writer, airconfig *Config) error {
	if o.Name == "" && !o.CurrentContext {
		return getContexts(out, airconfig)
	}
	return getContext(o, out, airconfig)
}

func getContext(o *ContextOptions, out io.Writer, airconfig *Config) error {
	cName := o.Name
	if o.CurrentContext {
		cName = airconfig.CurrentContext
	}
	context, err := airconfig.GetContext(cName)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s", context.PrettyString())
	return nil
}

func getContexts(out io.Writer, airconfig *Config) error {
	contexts, err := airconfig.GetContexts()
	if err != nil {
		return err
	}
	if len(contexts) == 0 {
		fmt.Fprintln(out, "No Contexts found in the configuration.")
	}
	for _, context := range contexts {
		fmt.Fprintf(out, "%s", context.PrettyString())
	}
	return nil
}

func RunSetAuthInfo(o *AuthInfoOptions, airconfig *Config, writeToStorage bool) (bool, error) {
	authinfoWasModified := false
	err := o.Validate()
	if err != nil {
		return authinfoWasModified, err
	}

	authinfoIWant := o.Name
	authinfo, err := airconfig.GetAuthInfo(authinfoIWant)
	if err != nil {
		var cerr ErrMissingConfig
		if !errors.As(err, &cerr) {
			// An error occurred, but it wasn't a "missing" config error.
			return authinfoWasModified, err
		}

		// authinfo didn't exist, create it
		// ignoring the returned added authinfo
		airconfig.AddAuthInfo(o)
	} else {
		// AuthInfo exists, lets update
		airconfig.ModifyAuthInfo(authinfo, o)
		authinfoWasModified = true
	}
	// Update configuration file just in time persistence approach
	if writeToStorage {
		if err := airconfig.PersistConfig(); err != nil {
			// Error that it didnt persist the changes
			return authinfoWasModified, ErrConfigFailed{}
		}
	}

	return authinfoWasModified, nil
}

func RunSetCluster(o *ClusterOptions, airconfig *Config, writeToStorage bool) (bool, error) {
	clusterWasModified := false
	err := o.Validate()
	if err != nil {
		return clusterWasModified, err
	}

	cluster, err := airconfig.GetCluster(o.Name, o.ClusterType)
	if err != nil {
		var cerr ErrMissingConfig
		if !errors.As(err, &cerr) {
			// An error occurred, but it wasn't a "missing" config error.
			return clusterWasModified, err
		}

		// Cluster didn't exist, create it
		_, err := airconfig.AddCluster(o)
		if err != nil {
			return clusterWasModified, err
		}
		clusterWasModified = false
	} else {
		// Cluster exists, lets update
		_, err := airconfig.ModifyCluster(cluster, o)
		if err != nil {
			return clusterWasModified, err
		}
		clusterWasModified = true
	}

	// Update configuration file
	// Just in time persistence approach
	if writeToStorage {
		if err := airconfig.PersistConfig(); err != nil {
			// Some warning here , that it didnt persist the changes because of this
			// Or should we float this up
			// What would it mean? No value.
			return clusterWasModified, err
		}
	}

	return clusterWasModified, nil
}

func RunSetContext(o *ContextOptions, airconfig *Config, writeToStorage bool) (bool, error) {
	contextWasModified := false
	err := o.Validate()
	if err != nil {
		return contextWasModified, err
	}

	contextIWant := o.Name
	context, err := airconfig.GetContext(contextIWant)
	if err != nil {
		var cerr ErrMissingConfig
		if !errors.As(err, &cerr) {
			// An error occurred, but it wasn't a "missing" config error.
			return contextWasModified, err
		}

		if o.CurrentContext {
			return contextWasModified, ErrMissingConfig{}
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
		contextWasModified = true
	}
	// Update configuration file just in time persistence approach
	if writeToStorage {
		if err := airconfig.PersistConfig(); err != nil {
			// Error that it didnt persist the changes
			return contextWasModified, ErrConfigFailed{}
		}
	}

	return contextWasModified, nil
}
