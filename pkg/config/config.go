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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"sigs.k8s.io/yaml"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"opendev.org/airship/airshipctl/pkg/util"
)

// LoadConfig populates the Config object using the files found at
// airshipConfigPath and kubeConfigPath
func (c *Config) LoadConfig(airshipConfigPath, kubeConfigPath string) error {
	err := c.loadFromAirConfig(airshipConfigPath)
	if err != nil {
		return err
	}

	err = c.loadKubeConfig(kubeConfigPath)
	if err != nil {
		return err
	}

	// Lets navigate through the kubeconfig to populate the references in airship config
	return c.reconcileConfig()
}

// loadFromAirConfig populates the Config from the file found at airshipConfigPath.
// If there is no file at airshipConfigPath, this function does nothing.
// An error is returned if:
// * airshipConfigPath is the empty string
// * the file at airshipConfigPath is inaccessible
// * the file at airshipConfigPath cannot be marshaled into Config
func (c *Config) loadFromAirConfig(airshipConfigPath string) error {
	if airshipConfigPath == "" {
		return errors.New("Configuration file location was not provided.")
	}

	// Remember where I loaded the Config from
	c.loadedConfigPath = airshipConfigPath

	// If I can read from the file, load from it
	if _, err := os.Stat(airshipConfigPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return util.ReadYAMLFile(airshipConfigPath, c)
}

func (c *Config) loadKubeConfig(kubeConfigPath string) error {
	// Will need this for persisting the changes
	c.kubeConfigPath = kubeConfigPath

	// If I can read from the file, load from it
	var err error
	if _, err = os.Stat(kubeConfigPath); os.IsNotExist(err) {
		// Default kubeconfig matching Airship target cluster
		c.kubeConfig = &clientcmdapi.Config{
			Clusters: map[string]*clientcmdapi.Cluster{
				AirshipDefaultContext: {
					Server: "https://172.17.0.1:6443",
				},
			},
			AuthInfos: map[string]*clientcmdapi.AuthInfo{
				"admin": {
					Username: "airship-admin",
				},
			},
			Contexts: map[string]*clientcmdapi.Context{
				AirshipDefaultContext: {
					Cluster:  AirshipDefaultContext,
					AuthInfo: "admin",
				},
			},
		}
		return nil
	} else if err != nil {
		return err
	}

	c.kubeConfig, err = clientcmd.LoadFromFile(kubeConfigPath)
	return err
}

// reconcileConfig serves two functions:
// 1 - it will consume from kubeconfig and update airship config
//     	For cluster that do not comply with the airship cluster type expectations a default
//	behavior will be implemented. Such as ,by default they will be tar or ephemeral
// 2 - it will update kubeconfig cluster objects with the appropriate <clustername>_<clustertype> convention
func (c *Config) reconcileConfig() error {
	updatedClusterNames, persistIt := c.reconcileClusters()
	c.reconcileContexts(updatedClusterNames)
	c.reconcileAuthInfos()
	c.reconcileCurrentContext()

	// I changed things during the reconciliation
	// Lets reflect them in the config files
	// Specially useful if the config is loaded during a get operation
	// If it was a Set this would have happened eventually any way
	if persistIt {
		return c.PersistConfig()
	}
	return nil
}

// reconcileClusters synchronizes the airshipconfig file with the kubeconfig file.
//
// It iterates over the clusters listed in the kubeconfig. If any cluster in
// the kubeconfig does not meet the <name>_<type> convention, the name is
// first changed to the airship default.
//
// It then updates the airshipconfig's names of those clusters, as well as the
// pointer to the clusters.
// If the cluster wasn't referenced prior to the call, it is created; otherwise
// it is modified.
//
// Finally, any clusters listed in the airshipconfig that are no longer
// referenced in the kubeconfig are deleted
//
// The function returns a mapping of changed names in the kubeconfig, as well
// as a boolean denoting that the config files need to be written to file
func (c *Config) reconcileClusters() (map[string]string, bool) {
	// updatedClusterNames is a mapping from OLD cluster names to NEW
	// cluster names. This will be used later when we update contexts
	updatedClusterNames := map[string]string{}

	persistIt := false
	for clusterName, cluster := range c.kubeConfig.Clusters {
		clusterComplexName := NewClusterComplexName()
		clusterComplexName.FromName(clusterName)
		// Check if the cluster from the kubeconfig file complies with
		// the airship naming convention
		if !clusterComplexName.validName() {
			clusterComplexName.SetDefaultType()
			// Update the kubeconfig with proper airship name
			c.kubeConfig.Clusters[clusterComplexName.Name()] = cluster
			delete(c.kubeConfig.Clusters, clusterName)

			// We also need to save the mapping from the old name
			// so we can update the context in the kubeconfig later
			updatedClusterNames[clusterName] = clusterComplexName.Name()

			// Since we've modified the kubeconfig object, we'll
			// need to let the caller know that the kubeconfig file
			// needs to be updated
			persistIt = true

			// Otherwise this is a cluster that didnt have an
			// airship cluster type, however when you added the
			// cluster type
			// Probable should just add a number _<COUNTER to it
		}

		// Update the airship config file
		if c.Clusters[clusterComplexName.ClusterName()] == nil {
			c.Clusters[clusterComplexName.ClusterName()] = NewClusterPurpose()
		}
		if c.Clusters[clusterComplexName.ClusterName()].ClusterTypes[clusterComplexName.ClusterType()] == nil {
			c.Clusters[clusterComplexName.ClusterName()].ClusterTypes[clusterComplexName.ClusterType()] = NewCluster()
		}
		configCluster := c.Clusters[clusterComplexName.ClusterName()].ClusterTypes[clusterComplexName.ClusterType()]
		if configCluster.NameInKubeconf != clusterComplexName.Name() {
			configCluster.NameInKubeconf = clusterComplexName.Name()
			// TODO What do we do with the BOOTSTRAP CONFIG
		}
		// Store the reference to the KubeConfig Cluster in the Airship Config
		configCluster.SetKubeCluster(cluster)
	}

	persistIt = c.rmConfigClusterStragglers(persistIt)

	return updatedClusterNames, persistIt
}

// Removes Cluster configuration that exist in Airship Config and do not have
// any kubeconfig appropriate <clustername>_<clustertype> entries
func (c *Config) rmConfigClusterStragglers(persistIt bool) bool {
	rccs := persistIt
	// Checking if there is any Cluster reference in airship config that does not match
	// an actual Cluster struct in kubeconfig
	for clusterName := range c.Clusters {
		for cType, cluster := range c.Clusters[clusterName].ClusterTypes {
			if _, found := c.kubeConfig.Clusters[cluster.NameInKubeconf]; !found {
				// Instead of removing it , I could add a empty entry in kubeconfig as well
				// Will see what is more appropriate with use of Modules configuration
				delete(c.Clusters[clusterName].ClusterTypes, cType)

				// If that was the last cluster type, then we
				// should delete the cluster entry
				if len(c.Clusters[clusterName].ClusterTypes) == 0 {
					delete(c.Clusters, clusterName)
				}
				rccs = true
			}
		}
	}
	return rccs
}

func (c *Config) reconcileContexts(updatedClusterNames map[string]string) {
	for key, context := range c.kubeConfig.Contexts {
		// Check if the Cluster name referred to by the context
		// was updated during the cluster reconcile
		if newName, ok := updatedClusterNames[context.Cluster]; ok {
			context.Cluster = newName
		}

		if c.Contexts[key] == nil {
			c.Contexts[key] = NewContext()
		}
		// Make sure the name matches
		c.Contexts[key].NameInKubeconf = context.Cluster
		c.Contexts[key].SetKubeContext(context)

		// What about if a Context refers to a properly named cluster
		// that does not exist in airship config
		clusterName := NewClusterComplexName()
		clusterName.FromName(context.Cluster)
		if clusterName.validName() && c.Clusters[clusterName.ClusterName()] == nil {
			// I cannot create this cluster, it will have empty information
			// Best course of action is to delete it I think
			delete(c.kubeConfig.Contexts, key)
		}
	}
	// Checking if there is any Context reference in airship config that does not match
	// an actual Context struct in kubeconfig, if they do not exists I will delete
	// Since context in airship config are only references mainly.
	for key := range c.Contexts {
		if c.kubeConfig.Contexts[key] == nil {
			delete(c.Contexts, key)
		}
	}
}

func (c *Config) reconcileAuthInfos() {
	for key, authinfo := range c.kubeConfig.AuthInfos {
		// Simple check if the AuthInfo name is referenced in airship config
		if c.AuthInfos[key] == nil && authinfo != nil {
			// Add the reference
			c.AuthInfos[key] = NewAuthInfo()
		}
		c.AuthInfos[key].SetKubeAuthInfo(authinfo)
	}
	// Checking if there is any AuthInfo reference in airship config that does not match
	// an actual Auth Info struct in kubeconfig
	for key := range c.AuthInfos {
		if c.kubeConfig.AuthInfos[key] == nil {
			delete(c.AuthInfos, key)
		}
	}
}

func (c *Config) reconcileCurrentContext() {
	// If the Airship current context is different that the current context in the kubeconfig
	// then
	//  - if the airship current context is valid, then updated kubeconfig CC
	//  - if the airship currentcontext is invalid, and the kubeconfig CC is valid, then create the reference
	//  - otherwise , they are both empty. Make sure

	if c.Contexts[c.CurrentContext] == nil { // Its not valid
		if c.Contexts[c.kubeConfig.CurrentContext] != nil {
			c.CurrentContext = c.kubeConfig.CurrentContext
		}
	} else {
		// Overpowers kubeConfig CurrentContext
		if c.kubeConfig.CurrentContext != c.CurrentContext {
			c.kubeConfig.CurrentContext = c.CurrentContext
		}
	}
}

// EnsureComplete verifies that a Config object is ready to use.
// A complete Config object meets the following criteria:
//   * At least 1 Cluster is defined
//   * At least 1 AuthInfo (user) is defined
//   * At least 1 Context is defined
//   * At least 1 Manifest is defined
//   * The CurrentContext is set
//   * The CurrentContext identifies an existing Context
//   * The CurrentContext identifies an existing Manifest
func (c *Config) EnsureComplete() error {
	if len(c.Clusters) == 0 {
		return ErrMissingConfig{
			What: "At least one cluster needs to be defined",
		}
	}

	if len(c.AuthInfos) == 0 {
		return ErrMissingConfig{
			What: "At least one Authentication Information (User) needs to be defined",
		}
	}

	if len(c.Contexts) == 0 {
		return ErrMissingConfig{
			What: "At least one Context needs to be defined",
		}
	}

	if len(c.Manifests) == 0 {
		return ErrMissingConfig{
			What: "At least one Manifest needs to be defined",
		}
	}

	if c.CurrentContext == "" {
		return ErrMissingConfig{
			What: "Current Context is not defined",
		}
	}

	currentContext, found := c.Contexts[c.CurrentContext]
	if !found {
		return ErrMissingConfig{
			What: fmt.Sprintf("Current Context (%s) does not identify a defined Context", c.CurrentContext),
		}
	}

	if _, found := c.Manifests[currentContext.Manifest]; !found {
		return ErrMissingConfig{
			What: fmt.Sprintf("Current Context (%s) does not identify a defined Manifest", c.CurrentContext),
		}
	}

	return nil
}

// PersistConfig updates the airshipctl config and kubeconfig files to match
// the current Config and KubeConfig objects.
// If either file did not previously exist, the file will be created.
// Otherwise, the file will be overwritten
func (c *Config) PersistConfig() error {
	airshipConfigYaml, err := c.ToYaml()
	if err != nil {
		return err
	}

	// WriteFile doesn't create the directory, create it if needed
	configDir := filepath.Dir(c.loadedConfigPath)
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return err
	}

	// Write the Airship Config file
	err = ioutil.WriteFile(c.loadedConfigPath, airshipConfigYaml, 0644)
	if err != nil {
		return err
	}

	// Persist the kubeconfig file referenced
	if err := clientcmd.WriteToFile(*c.kubeConfig, c.kubeConfigPath); err != nil {
		return err
	}

	return nil
}

func (c *Config) String() string {
	yamlData, err := c.ToYaml()
	// This is hiding the error perhaps
	if err != nil {
		return ""
	}
	return string(yamlData)
}

func (c *Config) ToYaml() ([]byte, error) {
	return yaml.Marshal(&c)
}

func (c *Config) LoadedConfigPath() string {
	return c.loadedConfigPath
}
func (c *Config) SetLoadedConfigPath(lcp string) {
	c.loadedConfigPath = lcp
}

func (c *Config) KubeConfigPath() string {
	return c.kubeConfigPath
}

func (c *Config) SetKubeConfigPath(kubeConfigPath string) {
	c.kubeConfigPath = kubeConfigPath
}

func (c *Config) KubeConfig() *clientcmdapi.Config {
	return c.kubeConfig
}

// Get A Cluster
func (c *Config) GetCluster(cName, cType string) (*Cluster, error) {
	_, exists := c.Clusters[cName]
	if !exists {
		return nil, ErrMissingConfig{What: fmt.Sprintf("Cluster with name '%s' of type '%s'", cName, cType)}
	}
	// Alternative to this would be enhance Cluster.String() to embed the appropriate kubeconfig cluster information
	cluster, exists := c.Clusters[cName].ClusterTypes[cType]
	if !exists {
		return nil, ErrMissingConfig{What: fmt.Sprintf("Cluster with name '%s' of type '%s'", cName, cType)}
	}
	return cluster, nil
}

func (c *Config) AddCluster(theCluster *ClusterOptions) (*Cluster, error) {
	// Need to create new cluster placeholder
	// Get list of ClusterPurposes that match the theCluster.name
	// Cluster might exists, but ClusterPurpose should not
	_, exists := c.Clusters[theCluster.Name]
	if !exists {
		c.Clusters[theCluster.Name] = NewClusterPurpose()
	}
	// Create the new Airship config Cluster
	nCluster := NewCluster()
	c.Clusters[theCluster.Name].ClusterTypes[theCluster.ClusterType] = nCluster
	// Create a new Kubeconfig Cluster object as well
	kcluster := clientcmdapi.NewCluster()
	clusterName := NewClusterComplexName()
	clusterName.WithType(theCluster.Name, theCluster.ClusterType)
	nCluster.NameInKubeconf = clusterName.Name()
	nCluster.SetKubeCluster(kcluster)

	c.KubeConfig().Clusters[clusterName.Name()] = kcluster

	// Ok , I have initialized structs for the Cluster information
	// We can use Modify to populate the correct information
	return c.ModifyCluster(nCluster, theCluster)
}

func (c *Config) ModifyCluster(cluster *Cluster, theCluster *ClusterOptions) (*Cluster, error) {
	kcluster := cluster.KubeCluster()
	if kcluster == nil {
		return cluster, nil
	}
	if theCluster.Server != "" {
		kcluster.Server = theCluster.Server
	}
	if theCluster.InsecureSkipTLSVerify {
		kcluster.InsecureSkipTLSVerify = theCluster.InsecureSkipTLSVerify
		// Specifying insecur mode clears any certificate authority
		if kcluster.InsecureSkipTLSVerify {
			kcluster.CertificateAuthority = ""
			kcluster.CertificateAuthorityData = nil
		}
	}
	if theCluster.CertificateAuthority == "" {
		return cluster, nil
	}

	if theCluster.EmbedCAData {
		readData, err := ioutil.ReadFile(theCluster.CertificateAuthority)
		kcluster.CertificateAuthorityData = readData
		if err != nil {
			return cluster, err
		}
		kcluster.InsecureSkipTLSVerify = false
		kcluster.CertificateAuthority = ""
	} else {
		caPath, err := filepath.Abs(theCluster.CertificateAuthority)
		if err != nil {
			return cluster, err
		}
		kcluster.CertificateAuthority = caPath
		// Specifying a certificate authority file clears certificate authority data and insecure mode
		if caPath != "" {
			kcluster.InsecureSkipTLSVerify = false
			kcluster.CertificateAuthorityData = nil
		}
	}
	return cluster, nil
}

// GetClusters returns all of the clusters associated with the Config sorted by name
func (c *Config) GetClusters() []*Cluster {
	keys := make([]string, 0, len(c.Clusters))
	for name := range c.Clusters {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	clusters := make([]*Cluster, 0, len(c.Clusters))
	for _, name := range keys {
		for _, clusterType := range AllClusterTypes {
			cluster, exists := c.Clusters[name].ClusterTypes[clusterType]
			if exists {
				// If it doesn't exist, then there must not be
				// a cluster with this name/type combination.
				// This is expected behavior
				clusters = append(clusters, cluster)
			}
		}
	}
	return clusters
}

// Context Operations from Config point of view
// Get Context
func (c *Config) GetContext(cName string) (*Context, error) {
	context, exists := c.Contexts[cName]
	if !exists {
		return nil, ErrMissingConfig{What: fmt.Sprintf("Context with name '%s'", cName)}
	}
	return context, nil
}

// GetContexts returns all of the contexts associated with the Config sorted by name
func (c *Config) GetContexts() []*Context {
	keys := make([]string, 0, len(c.Contexts))
	for name := range c.Contexts {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	contexts := make([]*Context, 0, len(c.Contexts))
	for _, name := range keys {
		contexts = append(contexts, c.Contexts[name])
	}
	return contexts
}

func (c *Config) AddContext(theContext *ContextOptions) *Context {
	// Create the new Airship config context
	nContext := NewContext()
	c.Contexts[theContext.Name] = nContext
	// Create a new Kubeconfig Context object as well
	kContext := clientcmdapi.NewContext()
	nContext.NameInKubeconf = theContext.Name
	contextName := NewClusterComplexName()
	contextName.WithType(theContext.Name, theContext.ClusterType)

	nContext.SetKubeContext(kContext)
	c.KubeConfig().Contexts[theContext.Name] = kContext

	// Ok , I have initialized structs for the Context information
	// We can use Modify to populate the correct information
	c.ModifyContext(nContext, theContext)
	return nContext
}

func (c *Config) ModifyContext(context *Context, theContext *ContextOptions) {
	kContext := context.KubeContext()
	if kContext == nil {
		return
	}
	if theContext.Cluster != "" {
		kContext.Cluster = theContext.Cluster
	}
	if theContext.AuthInfo != "" {
		kContext.AuthInfo = theContext.AuthInfo
	}
	if theContext.Manifest != "" {
		context.Manifest = theContext.Manifest
	}
	if theContext.Namespace != "" {
		kContext.Namespace = theContext.Namespace
	}
}

// CurrentContext methods Returns the appropriate information for the current context
// Current Context holds labels for the approriate config objects
//      Cluster is the name of the cluster for this context
//      ClusterType is the name of the clustertype for this context, it should be a flag we pass to it??
//      AuthInfo is the name of the authInfo for this context
//      Manifest is the default manifest to be use with this context
// Purpose for this method is simplifying the current context information
func (c *Config) GetCurrentContext() (*Context, error) {
	if err := c.EnsureComplete(); err != nil {
		return nil, err
	}
	currentContext, err := c.GetContext(c.CurrentContext)
	if err != nil {
		// this should not happen since Ensure Complete checks for this
		return nil, err
	}
	return currentContext, nil
}
func (c *Config) CurrentContextCluster() (*Cluster, error) {
	currentContext, err := c.GetCurrentContext()
	if err != nil {
		return nil, err
	}
	clusterName := NewClusterComplexName()
	clusterName.FromName(currentContext.KubeContext().Cluster)

	return c.Clusters[clusterName.ClusterName()].ClusterTypes[currentContext.ClusterType()], nil
}

func (c *Config) CurrentContextAuthInfo() (*AuthInfo, error) {
	currentContext, err := c.GetCurrentContext()
	if err != nil {
		return nil, err
	}

	return c.AuthInfos[currentContext.KubeContext().AuthInfo], nil
}
func (c *Config) CurrentContextManifest() (*Manifest, error) {
	currentContext, err := c.GetCurrentContext()
	if err != nil {
		return nil, err
	}

	return c.Manifests[currentContext.Manifest], nil
}

// Credential or AuthInfo related methods
func (c *Config) GetAuthInfo(aiName string) (*AuthInfo, error) {
	authinfo, exists := c.AuthInfos[aiName]
	if !exists {
		return nil, ErrMissingConfig{What: fmt.Sprintf("User credentials with name '%s'", aiName)}
	}
	return authinfo, nil
}

// GetAuthInfos returns a slice containing all the AuthInfos associated with
// the Config sorted by name
func (c *Config) GetAuthInfos() []*AuthInfo {
	keys := make([]string, 0, len(c.AuthInfos))
	for name := range c.AuthInfos {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	authInfos := make([]*AuthInfo, 0, len(c.AuthInfos))
	for _, name := range keys {
		authInfos = append(authInfos, c.AuthInfos[name])
	}
	return authInfos
}

func (c *Config) AddAuthInfo(theAuthInfo *AuthInfoOptions) *AuthInfo {
	// Create the new Airship config context
	nAuthInfo := NewAuthInfo()
	c.AuthInfos[theAuthInfo.Name] = nAuthInfo
	// Create a new Kubeconfig AuthInfo object as well
	kAuthInfo := clientcmdapi.NewAuthInfo()
	nAuthInfo.SetKubeAuthInfo(kAuthInfo)
	c.KubeConfig().AuthInfos[theAuthInfo.Name] = kAuthInfo

	c.ModifyAuthInfo(nAuthInfo, theAuthInfo)
	return nAuthInfo
}

func (c *Config) ModifyAuthInfo(authinfo *AuthInfo, theAuthInfo *AuthInfoOptions) {
	kAuthInfo := authinfo.KubeAuthInfo()
	if kAuthInfo == nil {
		return
	}
	if theAuthInfo.ClientCertificate != "" {
		kAuthInfo.ClientCertificate = theAuthInfo.ClientCertificate
	}
	if theAuthInfo.Token != "" {
		kAuthInfo.Token = theAuthInfo.Token
	}
	if theAuthInfo.Username != "" {
		kAuthInfo.Username = theAuthInfo.Username
	}
	if theAuthInfo.Password != "" {
		kAuthInfo.Password = theAuthInfo.Password
	}
	if theAuthInfo.ClientKey != "" {
		kAuthInfo.ClientKey = theAuthInfo.ClientKey
	}
}

// CurrentContextBootstrapInfo returns bootstrap info for current context
func (c *Config) CurrentContextBootstrapInfo() (*Bootstrap, error) {
	currentCluster, err := c.CurrentContextCluster()
	if err != nil {
		return nil, err
	}

	bootstrap, exists := c.ModulesConfig.BootstrapInfo[currentCluster.Bootstrap]
	if !exists {
		return nil, ErrBootstrapInfoNotFound{Name: currentCluster.Bootstrap}
	}
	return bootstrap, nil
}

// Purge removes the config file
func (c *Config) Purge() error {
	return os.Remove(c.loadedConfigPath)
}

func (c *Config) Equal(d *Config) bool {
	if d == nil {
		return d == c
	}
	clusterEq := reflect.DeepEqual(c.Clusters, d.Clusters)
	authInfoEq := reflect.DeepEqual(c.AuthInfos, d.AuthInfos)
	contextEq := reflect.DeepEqual(c.Contexts, d.Contexts)
	manifestEq := reflect.DeepEqual(c.Manifests, d.Manifests)
	modulesEq := reflect.DeepEqual(c.ModulesConfig, d.ModulesConfig)
	return c.Kind == d.Kind &&
		c.APIVersion == d.APIVersion &&
		clusterEq && authInfoEq && contextEq && manifestEq && modulesEq
}

// Cluster functions
func (c *Cluster) Equal(d *Cluster) bool {
	if d == nil {
		return d == c
	}
	return c.NameInKubeconf == d.NameInKubeconf &&
		c.Bootstrap == d.Bootstrap
}

func (c *Cluster) String() string {
	cyaml, err := yaml.Marshal(&c)
	if err != nil {
		return ""
	}
	kcluster := c.KubeCluster()
	kyaml, err := yaml.Marshal(&kcluster)
	if err != nil {
		return string(cyaml)
	}

	return fmt.Sprintf("%s\n%s", string(cyaml), string(kyaml))
}

func (c *Cluster) PrettyString() string {
	clusterName := NewClusterComplexName()
	clusterName.FromName(c.NameInKubeconf)

	return fmt.Sprintf("Cluster: %s\n%s:\n%s",
		clusterName.ClusterName(), clusterName.ClusterType(), c)
}

func (c *Cluster) KubeCluster() *clientcmdapi.Cluster {
	return c.kCluster
}
func (c *Cluster) SetKubeCluster(kc *clientcmdapi.Cluster) {
	c.kCluster = kc
}

// Context functions
func (c *Context) Equal(d *Context) bool {
	if d == nil {
		return d == c
	}
	return c.NameInKubeconf == d.NameInKubeconf &&
		c.Manifest == d.Manifest &&
		c.kContext == d.kContext
}

func (c *Context) String() string {
	cyaml, err := yaml.Marshal(&c)
	if err != nil {
		return ""
	}
	kcluster := c.KubeContext()
	kyaml, err := yaml.Marshal(&kcluster)
	if err != nil {
		return string(cyaml)
	}
	return fmt.Sprintf("%s\n%s", string(cyaml), string(kyaml))
}

func (c *Context) PrettyString() string {
	clusterName := NewClusterComplexName()
	clusterName.FromName(c.NameInKubeconf)

	return fmt.Sprintf("Context: %s\n%s\n",
		clusterName.ClusterName(), c.String())
}

func (c *Context) KubeContext() *clientcmdapi.Context {
	return c.kContext
}

func (c *Context) SetKubeContext(kc *clientcmdapi.Context) {
	c.kContext = kc
}

func (c *Context) ClusterType() string {
	clusterName := NewClusterComplexName()
	clusterName.FromName(c.NameInKubeconf)
	return clusterName.ClusterType()
}

// AuthInfo functions
func (c *AuthInfo) Equal(d *AuthInfo) bool {
	if d == nil {
		return c == d
	}
	return c.kAuthInfo == d.kAuthInfo
}

func (c *AuthInfo) String() string {
	kauthinfo := c.KubeAuthInfo()
	kyaml, err := yaml.Marshal(&kauthinfo)
	if err != nil {
		return ""
	}
	return string(kyaml)
}

func (c *AuthInfo) KubeAuthInfo() *clientcmdapi.AuthInfo {
	return c.kAuthInfo
}
func (c *AuthInfo) SetKubeAuthInfo(kc *clientcmdapi.AuthInfo) {
	c.kAuthInfo = kc
}

// Manifest functions
func (m *Manifest) Equal(n *Manifest) bool {
	if n == nil {
		return n == m
	}
	repositoryEq := reflect.DeepEqual(m.Repository, n.Repository)
	extraReposEq := reflect.DeepEqual(m.ExtraRepositories, n.ExtraRepositories)
	return repositoryEq && extraReposEq && m.TargetPath == n.TargetPath
}

func (m *Manifest) String() string {
	yamlData, err := yaml.Marshal(&m)
	if err != nil {
		return ""
	}
	return string(yamlData)
}

// Modules functions
func (m *Modules) Equal(n *Modules) bool {
	if n == nil {
		return n == m
	}
	return reflect.DeepEqual(m.BootstrapInfo, n.BootstrapInfo)
}
func (m *Modules) String() string {
	yamlData, err := yaml.Marshal(&m)
	if err != nil {
		return ""
	}
	return string(yamlData)
}

// Bootstrap functions
func (b *Bootstrap) Equal(c *Bootstrap) bool {
	if c == nil {
		return b == c
	}
	contEq := reflect.DeepEqual(b.Container, c.Container)
	bldrEq := reflect.DeepEqual(b.Builder, c.Builder)
	return contEq && bldrEq
}

func (b *Bootstrap) String() string {
	yamlData, err := yaml.Marshal(&b)
	if err != nil {
		return ""
	}
	return string(yamlData)
}

// Container functions
func (c *Container) Equal(d *Container) bool {
	if d == nil {
		return d == c
	}
	return c.Volume == d.Volume &&
		c.Image == d.Image &&
		c.ContainerRuntime == d.ContainerRuntime
}

func (c *Container) String() string {
	yamlData, err := yaml.Marshal(&c)
	if err != nil {
		return ""
	}
	return string(yamlData)
}

// Builder functions
func (b *Builder) Equal(c *Builder) bool {
	if c == nil {
		return b == c
	}
	return b.UserDataFileName == c.UserDataFileName &&
		b.NetworkConfigFileName == c.NetworkConfigFileName &&
		b.OutputMetadataFileName == c.OutputMetadataFileName
}

func (b *Builder) String() string {
	yamlData, err := yaml.Marshal(&b)
	if err != nil {
		return ""
	}
	return string(yamlData)
}

// ClusterComplexName functions
func (c *ClusterComplexName) validName() bool {
	err := ValidClusterType(c.clusterType)
	return c.clusterName != "" && err == nil
}

func (c *ClusterComplexName) FromName(clusterName string) {
	if clusterName == "" {
		return
	}

	userNameSplit := strings.Split(clusterName, AirshipClusterNameSep)
	if len(userNameSplit) == 1 {
		c.clusterName = clusterName
		return
	}

	for _, cType := range AllClusterTypes {
		if userNameSplit[len(userNameSplit)-1] == cType {
			c.clusterType = userNameSplit[len(userNameSplit)-1]
			c.clusterName = strings.Join(userNameSplit[:len(userNameSplit)-1], AirshipClusterNameSep)
			return
		}
	}
}

func (c *ClusterComplexName) WithType(clusterName string, clusterType string) {
	c.FromName(clusterName)
	c.SetClusterType(clusterType)
}
func (c *ClusterComplexName) Name() string {
	s := []string{c.clusterName, c.clusterType}
	return strings.Join(s, AirshipClusterNameSep)
}
func (c *ClusterComplexName) ClusterName() string {
	return c.clusterName
}

func (c *ClusterComplexName) ClusterType() string {
	return c.clusterType
}
func (c *ClusterComplexName) SetClusterName(cn string) {
	c.clusterName = cn
}

func (c *ClusterComplexName) SetClusterType(ct string) {
	c.clusterType = ct
}
func (c *ClusterComplexName) SetDefaultType() {
	c.SetClusterType(AirshipClusterDefaultType)
}
func (c *ClusterComplexName) String() string {
	return fmt.Sprintf("clusterName:%s, clusterType:%s", c.clusterName, c.clusterType)
}
func ValidClusterType(clusterType string) error {
	for _, validType := range AllClusterTypes {
		if clusterType == validType {
			return nil
		}
	}
	return fmt.Errorf("Cluster Type must be one of %v", AllClusterTypes)
}

/* ______________________________
PLACEHOLDER UNTIL I IDENTIFY if CLIENTADM
HAS SOMETHING LIKE THIS
*/

func KClusterString(kCluster *clientcmdapi.Cluster) string {
	yamlData, err := yaml.Marshal(&kCluster)
	if err != nil {
		return ""
	}

	return string(yamlData)
}

func KContextString(kContext *clientcmdapi.Context) string {
	yamlData, err := yaml.Marshal(&kContext)
	if err != nil {
		return ""
	}

	return string(yamlData)
}

func KAuthInfoString(kAuthInfo *clientcmdapi.AuthInfo) string {
	yamlData, err := yaml.Marshal(&kAuthInfo)
	if err != nil {
		return ""
	}

	return string(yamlData)
}
