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
	"path"
	"path/filepath"
	"sort"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/yaml"

	"opendev.org/airship/airshipctl/pkg/util"
)

// Where possible, json tags match the cli argument names.
// Top level config objects and all values required for proper functioning are not "omitempty".
// Any truly optional piece of config is allowed to be omitted.

// Config holds the information required by airshipctl commands
// It is somewhat a superset of what a kubeconfig looks like, we allow for this overlaps by providing
// a mechanism to consume or produce a kubeconfig into / from the airship config.
type Config struct {
	// +optional
	Kind string `json:"kind,omitempty"`

	// +optional
	APIVersion string `json:"apiVersion,omitempty"`

	// Clusters is a map of referenceable names to cluster configs
	Clusters map[string]*ClusterPurpose `json:"clusters"`

	// AuthInfos is a map of referenceable names to user configs
	AuthInfos map[string]*AuthInfo `json:"users"`

	// Contexts is a map of referenceable names to context configs
	Contexts map[string]*Context `json:"contexts"`

	// Manifests is a map of referenceable names to documents
	Manifests map[string]*Manifest `json:"manifests"`

	// CurrentContext is the name of the context that you would like to use by default
	CurrentContext string `json:"currentContext"`

	// Management configuration defines management information for all baremetal hosts in a cluster.
	ManagementConfiguration map[string]*ManagementConfiguration `json:"managementConfiguration"`

	// BootstrapInfo is the configuration for container runtime, ISO builder and remote management
	BootstrapInfo map[string]*Bootstrap `json:"bootstrapInfo"`

	// loadedConfigPath is the full path to the the location of the config
	// file from which this config was loaded
	// +not persisted in file
	loadedConfigPath string

	// kubeConfigPath is the full path to the the location of the
	// kubeconfig file associated with this airship config instance
	// +not persisted in file
	kubeConfigPath string

	// Private instance of Kube Config content as an object
	kubeConfig *clientcmdapi.Config
}

// LoadConfig populates the Config object using the files found at
// airshipConfigPath and kubeConfigPath
func (c *Config) LoadConfig(airshipConfigPath, kubeConfigPath string, create bool) error {
	err := c.loadFromAirConfig(airshipConfigPath, create)
	if err != nil {
		return err
	}

	err = c.loadKubeConfig(kubeConfigPath, create)
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
func (c *Config) loadFromAirConfig(airshipConfigPath string, create bool) error {
	if airshipConfigPath == "" {
		return errors.New("configuration file location was not provided")
	}

	// Remember where I loaded the Config from
	c.loadedConfigPath = airshipConfigPath

	// If I can read from the file, load from it
	// throw an error otherwise
	if _, err := os.Stat(airshipConfigPath); os.IsNotExist(err) && create {
		return nil
	} else if err != nil {
		return err
	}

	return util.ReadYAMLFile(airshipConfigPath, c)
}

func (c *Config) loadKubeConfig(kubeConfigPath string, create bool) error {
	// Will need this for persisting the changes
	c.kubeConfigPath = kubeConfigPath

	// If I can read from the file, load from it
	var err error
	if _, err = os.Stat(kubeConfigPath); os.IsNotExist(err) && create {
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
		clusterComplexName := NewClusterComplexNameFromKubeClusterName(clusterName)
		// Check if the cluster from the kubeconfig file complies with
		// the airship naming convention
		if clusterName != clusterComplexName.String() {
			// Update the kubeconfig with proper airship name
			c.kubeConfig.Clusters[clusterComplexName.String()] = cluster
			delete(c.kubeConfig.Clusters, clusterName)

			// We also need to save the mapping from the old name
			// so we can update the context in the kubeconfig later
			updatedClusterNames[clusterName] = clusterComplexName.String()

			// Since we've modified the kubeconfig object, we'll
			// need to let the caller know that the kubeconfig file
			// needs to be updated
			persistIt = true

			// Otherwise this is a cluster that didnt have an
			// airship cluster type, however when you added the
			// cluster type
			// Probable should just add a number _<COUNTER to it
		}

		// The cluster in the kubeconfig is not present in the airship config. Create it.
		if c.Clusters[clusterComplexName.Name] == nil {
			c.Clusters[clusterComplexName.Name] = NewClusterPurpose()
		}

		// NOTE(drewwalters96): This is a user error because a cluster is defined in name but incomplete. We
		// need to fail sooner than this function; add up-front validation for this later.
		if c.Clusters[clusterComplexName.Name].ClusterTypes == nil {
			c.Clusters[clusterComplexName.Name].ClusterTypes = make(map[string]*Cluster)
		}

		// The cluster is defined, but the type is not. Define the type.
		if c.Clusters[clusterComplexName.Name].ClusterTypes[clusterComplexName.Type] == nil {
			c.Clusters[clusterComplexName.Name].ClusterTypes[clusterComplexName.Type] = NewCluster()
		}

		// Point cluster at kubeconfig
		configCluster := c.Clusters[clusterComplexName.Name].ClusterTypes[clusterComplexName.Type]
		configCluster.NameInKubeconf = clusterComplexName.String()

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

		// What about if a Context refers to a cluster that does not
		// exist in airship config
		clusterName := NewClusterComplexNameFromKubeClusterName(context.Cluster)
		if c.Clusters[clusterName.Name] == nil {
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
		c.AuthInfos[key].authInfo = authinfo
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

// ToYaml returns a YAML document
// It serializes the given Config object to a valid YAML document
func (c *Config) ToYaml() ([]byte, error) {
	return yaml.Marshal(&c)
}

// LoadedConfigPath returns the file path of airship config
// from where the current Config object is created
func (c *Config) LoadedConfigPath() string {
	return c.loadedConfigPath
}

// SetLoadedConfigPath updates the file path of airship config
// in the Config object
func (c *Config) SetLoadedConfigPath(lcp string) {
	c.loadedConfigPath = lcp
}

// KubeConfigPath returns the file path of the kube config
// from Config object
func (c *Config) KubeConfigPath() string {
	return c.kubeConfigPath
}

// SetKubeConfigPath updates the file path of the kubeconfig
// in Config object
func (c *Config) SetKubeConfigPath(kubeConfigPath string) {
	c.kubeConfigPath = kubeConfigPath
}

// KubeConfig returns kube config object from the
// context of current Config object
func (c *Config) KubeConfig() *clientcmdapi.Config {
	return c.kubeConfig
}

// SetKubeConfig updates kube config in Config object
func (c *Config) SetKubeConfig(kubeConfig *clientcmdapi.Config) {
	c.kubeConfig = kubeConfig
}

// GetCluster returns a cluster instance
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

// AddCluster creates a new cluster and returns the
// newly created cluster object
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
	// Create a new KubeConfig Cluster object as well
	kcluster := clientcmdapi.NewCluster()
	clusterName := NewClusterComplexName(theCluster.Name, theCluster.ClusterType)
	nCluster.NameInKubeconf = clusterName.String()
	nCluster.SetKubeCluster(kcluster)

	c.KubeConfig().Clusters[clusterName.String()] = kcluster

	// Ok , I have initialized structs for the Cluster information
	// We can use Modify to populate the correct information
	return c.ModifyCluster(nCluster, theCluster)
}

// ModifyCluster updates cluster object with given cluster options
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

// GetContext returns a context instance
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

// GetManagementConfiguration retrieves a management configuration by name.
func (c *Config) GetManagementConfiguration(name string) (*ManagementConfiguration, error) {
	managementCfg, exists := c.ManagementConfiguration[name]
	if !exists {
		return nil, ErrManagementConfigurationNotFound{Name: name}
	}

	return managementCfg, nil
}

// AddContext creates a new context and returns the instance of
// newly created context
func (c *Config) AddContext(theContext *ContextOptions) *Context {
	// Create the new Airship config context
	nContext := NewContext()
	c.Contexts[theContext.Name] = nContext
	// Create a new KubeConfig Context object as well
	context := clientcmdapi.NewContext()
	nContext.NameInKubeconf = theContext.Name

	nContext.SetKubeContext(context)
	c.KubeConfig().Contexts[theContext.Name] = context

	// Ok , I have initialized structs for the Context information
	// We can use Modify to populate the correct information
	c.ModifyContext(nContext, theContext)
	return nContext
}

// ModifyContext updates Context object with given given context options
func (c *Config) ModifyContext(context *Context, theContext *ContextOptions) {
	kubeContext := context.KubeContext()
	if kubeContext == nil {
		return
	}
	if theContext.Cluster != "" {
		kubeContext.Cluster = theContext.Cluster
	}
	if theContext.AuthInfo != "" {
		kubeContext.AuthInfo = theContext.AuthInfo
	}
	if theContext.Manifest != "" {
		context.Manifest = theContext.Manifest
	}
	if theContext.Namespace != "" {
		kubeContext.Namespace = theContext.Namespace
	}
}

// GetCurrentContext methods Returns the appropriate information for the current context
// Current Context holds labels for the approriate config objects
//      Cluster is the name of the cluster for this context
//      ClusterType is the name of the clustertype for this context, it should be a flag we pass to it??
//      AuthInfo is the name of the authInfo for this context
//      Manifest is the default manifest to be use with this context
// Purpose for this method is simplifying the current context information
func (c *Config) GetCurrentContext() (*Context, error) {
	currentContext, err := c.GetContext(c.CurrentContext)
	if err != nil {
		// this should not happen since Ensure Complete checks for this
		return nil, err
	}
	return currentContext, nil
}

// CurrentContextCluster returns the Cluster for the current context
func (c *Config) CurrentContextCluster() (*Cluster, error) {
	currentContext, err := c.GetCurrentContext()
	if err != nil {
		return nil, err
	}
	clusterName := NewClusterComplexNameFromKubeClusterName(currentContext.KubeContext().Cluster)

	return c.Clusters[clusterName.Name].ClusterTypes[currentContext.ClusterType()], nil
}

// CurrentContextAuthInfo returns the AuthInfo for the current context
func (c *Config) CurrentContextAuthInfo() (*AuthInfo, error) {
	currentContext, err := c.GetCurrentContext()
	if err != nil {
		return nil, err
	}

	return c.AuthInfos[currentContext.KubeContext().AuthInfo], nil
}

// CurrentContextManifest returns the manifest for the current context
func (c *Config) CurrentContextManifest() (*Manifest, error) {
	currentContext, err := c.GetCurrentContext()
	if err != nil {
		return nil, err
	}

	return c.Manifests[currentContext.Manifest], nil
}

// CurrentContextEntryPoint returns path to build bundle based on clusterType and phase
// example CurrentContextEntryPoint("ephemeral", "initinfra")
func (c *Config) CurrentContextEntryPoint(phase string) (string, error) {
	clusterType, err := c.CurrentContextClusterType()
	if err != nil {
		return "", err
	}

	err = ValidClusterType(clusterType)
	if err != nil {
		return "", err
	}
	ccm, err := c.CurrentContextManifest()
	if err != nil {
		return "", err
	}
	_, exists := ccm.Repositories[ccm.PrimaryRepositoryName]
	if !exists {
		return "", ErrMissingPrimaryRepo{}
	}
	epp := path.Join(ccm.TargetPath, ccm.SubPath, clusterType, phase)
	if _, err := os.Stat(epp); err != nil {
		return "", ErrMissingPhaseDocument{PhaseName: phase}
	}
	return epp, nil
}

// CurrentContextTargetPath returns target path from current context's manifest
func (c *Config) CurrentContextTargetPath() (string, error) {
	ccm, err := c.CurrentContextManifest()
	if err != nil {
		return "", err
	}
	return ccm.TargetPath, nil
}

// CurrentContextClusterType returns cluster type of current context
func (c *Config) CurrentContextClusterType() (string, error) {
	context, err := c.GetCurrentContext()
	if err != nil {
		return "", err
	}
	return context.ClusterType(), nil
}

// CurrentContextClusterName returns cluster name of current context
func (c *Config) CurrentContextClusterName() (string, error) {
	context, err := c.GetCurrentContext()
	if err != nil {
		return "", err
	}
	return context.ClusterName(), nil
}

// GetAuthInfo returns an instance of authino
// Credential or AuthInfo related methods
func (c *Config) GetAuthInfo(aiName string) (*AuthInfo, error) {
	authinfo, exists := c.AuthInfos[aiName]
	if !exists {
		return nil, ErrMissingConfig{What: fmt.Sprintf("User credentials with name '%s'", aiName)}
	}
	decodedAuthInfo, err := DecodeAuthInfo(authinfo.authInfo)
	if err != nil {
		return nil, err
	}
	authinfo.authInfo = decodedAuthInfo
	return authinfo, nil
}

// GetAuthInfos returns a slice containing all the AuthInfos associated with
// the Config sorted by name
func (c *Config) GetAuthInfos() ([]*AuthInfo, error) {
	keys := make([]string, 0, len(c.AuthInfos))
	for name := range c.AuthInfos {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	authInfos := make([]*AuthInfo, 0, len(c.AuthInfos))
	for _, name := range keys {
		decodedAuthInfo, err := DecodeAuthInfo(c.AuthInfos[name].authInfo)
		if err != nil {
			return []*AuthInfo{}, err
		}
		c.AuthInfos[name].authInfo = decodedAuthInfo
		authInfos = append(authInfos, c.AuthInfos[name])
	}
	return authInfos, nil
}

// AddAuthInfo creates new AuthInfo with context details updated
// in the  airship config and kube config
func (c *Config) AddAuthInfo(theAuthInfo *AuthInfoOptions) *AuthInfo {
	// Create the new Airship config context
	nAuthInfo := NewAuthInfo()
	c.AuthInfos[theAuthInfo.Name] = nAuthInfo
	// Create a new KubeConfig AuthInfo object as well
	authInfo := clientcmdapi.NewAuthInfo()
	nAuthInfo.authInfo = authInfo
	c.KubeConfig().AuthInfos[theAuthInfo.Name] = authInfo

	c.ModifyAuthInfo(nAuthInfo, theAuthInfo)
	return nAuthInfo
}

// ModifyAuthInfo updates the AuthInfo in the Config object
func (c *Config) ModifyAuthInfo(authinfo *AuthInfo, theAuthInfo *AuthInfoOptions) {
	kubeAuthInfo := EncodeAuthInfo(authinfo.KubeAuthInfo())
	if kubeAuthInfo == nil {
		return
	}
	if theAuthInfo.ClientCertificate != "" {
		kubeAuthInfo.ClientCertificate = EncodeString(theAuthInfo.ClientCertificate)
	}
	if theAuthInfo.Token != "" {
		kubeAuthInfo.Token = EncodeString(theAuthInfo.Token)
	}
	if theAuthInfo.Username != "" {
		kubeAuthInfo.Username = theAuthInfo.Username
	}
	if theAuthInfo.Password != "" {
		kubeAuthInfo.Password = EncodeString(theAuthInfo.Password)
	}
	if theAuthInfo.ClientKey != "" {
		kubeAuthInfo.ClientKey = EncodeString(theAuthInfo.ClientKey)
	}
}

// ImportFromKubeConfig absorbs the clusters, contexts and credentials from the
// given kubeConfig
func (c *Config) ImportFromKubeConfig(kubeConfigPath string) error {
	_, err := os.Stat(kubeConfigPath)
	if err != nil {
		return err
	}

	kubeConfig, err := clientcmd.LoadFromFile(kubeConfigPath)
	if err != nil {
		return err
	}
	c.importClusters(kubeConfig)
	c.importContexts(kubeConfig)
	c.importAuthInfos(kubeConfig)
	return c.PersistConfig()
}

func (c *Config) importClusters(importKubeConfig *clientcmdapi.Config) {
	for clusterName, cluster := range importKubeConfig.Clusters {
		clusterComplexName := NewClusterComplexNameFromKubeClusterName(clusterName)
		if _, err := c.GetCluster(clusterComplexName.Name, clusterComplexName.Type); err == nil {
			// err == nil implies that we were successfully able to
			// get the cluster from the existing configuration.
			// Since existing clusters takes precedence, skip this cluster
			continue
		}

		// Initialize the new cluster for the airship configuration
		airshipCluster := NewCluster()
		airshipCluster.NameInKubeconf = clusterComplexName.String()
		// Store the reference to the KubeConfig Cluster in the Airship Config
		airshipCluster.SetKubeCluster(cluster)

		// Update the airship configuration
		if _, ok := c.Clusters[clusterComplexName.Name]; !ok {
			c.Clusters[clusterComplexName.Name] = NewClusterPurpose()
		}
		c.Clusters[clusterComplexName.Name].ClusterTypes[clusterComplexName.Type] = airshipCluster
		c.kubeConfig.Clusters[clusterComplexName.String()] = cluster
	}
}

func (c *Config) importContexts(importKubeConfig *clientcmdapi.Config) {
	// TODO(howell): This function doesn't handle the case when an incoming
	// context refers to a cluster that doesn't exist in the airship
	// configuration.
	for kubeContextName, kubeContext := range importKubeConfig.Contexts {
		if _, ok := c.kubeConfig.Contexts[kubeContextName]; ok {
			// Since existing contexts take precedence, skip this context
			continue
		}

		clusterComplexName := NewClusterComplexNameFromKubeClusterName(kubeContext.Cluster)
		if kubeContext.Cluster != clusterComplexName.String() {
			// If the name of cluster from the kubeConfig doesn't
			// match the clusterComplexName, it needs to be updated
			kubeContext.Cluster = clusterComplexName.String()
		}

		airshipContext, ok := c.Contexts[kubeContextName]
		if !ok {
			airshipContext = NewContext()
		}
		airshipContext.NameInKubeconf = kubeContext.Cluster
		airshipContext.Manifest = AirshipDefaultManifest
		airshipContext.SetKubeContext(kubeContext)

		// Store the contexts in the airship configuration
		c.Contexts[kubeContextName] = airshipContext
		c.kubeConfig.Contexts[kubeContextName] = kubeContext
	}
}

func (c *Config) importAuthInfos(importKubeConfig *clientcmdapi.Config) {
	for key, authinfo := range importKubeConfig.AuthInfos {
		if _, ok := c.AuthInfos[key]; ok {
			// Since existing credentials take precedence, skip this credential
			continue
		}

		c.AuthInfos[key] = NewAuthInfo()
		c.AuthInfos[key].SetKubeAuthInfo(authinfo)
		c.kubeConfig.AuthInfos[key] = authinfo
	}
}

// CurrentContextBootstrapInfo returns bootstrap info for current context
func (c *Config) CurrentContextBootstrapInfo() (*Bootstrap, error) {
	currentCluster, err := c.CurrentContextCluster()
	if err != nil {
		return nil, err
	}

	if currentCluster.Bootstrap == "" {
		return nil, ErrMissingConfig{
			What: fmt.Sprintf("No bootstrapInfo defined for context %q", c.CurrentContext),
		}
	}

	bootstrap, exists := c.BootstrapInfo[currentCluster.Bootstrap]
	if !exists {
		return nil, ErrBootstrapInfoNotFound{Name: currentCluster.Bootstrap}
	}
	return bootstrap, nil
}

// GetManifests returns all of the Manifests associated with the Config sorted by name
func (c *Config) GetManifests() []*Manifest {
	keys := make([]string, 0, len(c.Manifests))
	for name := range c.Manifests {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	manifests := make([]*Manifest, 0, len(c.Manifests))
	for _, name := range keys {
		manifests = append(manifests, c.Manifests[name])
	}
	return manifests
}

// AddManifest creates new Manifest
func (c *Config) AddManifest(theManifest *ManifestOptions) *Manifest {
	nManifest := NewManifest()
	c.Manifests[theManifest.Name] = nManifest
	err := c.ModifyManifest(nManifest, theManifest)
	if err != nil {
		return nil
	}
	return nManifest
}

// ModifyManifest set actual values to manifests
func (c *Config) ModifyManifest(manifest *Manifest, theManifest *ManifestOptions) error {
	if theManifest.IsPrimary {
		manifest.PrimaryRepositoryName = theManifest.RepoName
	}
	if theManifest.SubPath != "" {
		manifest.SubPath = theManifest.SubPath
	}
	if theManifest.TargetPath != "" {
		manifest.TargetPath = theManifest.TargetPath
	}
	// There is no repository details to be updated
	if theManifest.RepoName == "" {
		return nil
	}
	//when setting an existing repository as primary, verify whether the repository exists
	//and user is also not passing any repository URL
	if theManifest.IsPrimary && theManifest.URL == "" && (manifest.Repositories[theManifest.RepoName] == nil) {
		return ErrRepositoryNotFound{theManifest.RepoName}
	}
	repository, exists := manifest.Repositories[theManifest.RepoName]
	if !exists {
		_, err := c.AddRepository(manifest, theManifest)
		if err != nil {
			return err
		}
	} else {
		err := c.ModifyRepository(repository, theManifest)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddRepository creates new Repository
func (c *Config) AddRepository(manifest *Manifest, theManifest *ManifestOptions) (*Repository, error) {
	nRepository := NewRepository()
	manifest.Repositories[theManifest.RepoName] = nRepository
	err := c.ModifyRepository(nRepository, theManifest)
	if err != nil {
		return nil, err
	}
	return nRepository, nil
}

// ModifyRepository set actual values to repository
func (c *Config) ModifyRepository(repository *Repository, theManifest *ManifestOptions) error {
	if theManifest.URL != "" {
		repository.URLString = theManifest.URL
	}
	if theManifest.Branch != "" {
		repository.CheckoutOptions.Branch = theManifest.Branch
	}
	if theManifest.CommitHash != "" {
		repository.CheckoutOptions.CommitHash = theManifest.CommitHash
	}
	if theManifest.Tag != "" {
		repository.CheckoutOptions.Tag = theManifest.Tag
	}
	if theManifest.Force {
		repository.CheckoutOptions.ForceCheckout = theManifest.Force
	}
	possibleValues := [3]string{repository.CheckoutOptions.CommitHash,
		repository.CheckoutOptions.Branch, repository.CheckoutOptions.Tag}
	var count int
	for _, val := range possibleValues {
		if val != "" {
			count++
		}
	}
	if count > 1 {
		return ErrMutuallyExclusiveCheckout{}
	}
	if count == 0 {
		return ErrMissingRepoCheckoutOptions{}
	}
	return nil
}

// CurrentContextManagementConfig returns the management options for the current context
func (c *Config) CurrentContextManagementConfig() (*ManagementConfiguration, error) {
	currentCluster, err := c.CurrentContextCluster()
	if err != nil {
		return nil, err
	}

	if currentCluster.ManagementConfiguration == "" {
		return nil, ErrMissingConfig{
			What: fmt.Sprintf("No management config listed for cluster %s", currentCluster.NameInKubeconf),
		}
	}

	managementCfg, exists := c.ManagementConfiguration[currentCluster.ManagementConfiguration]
	if !exists {
		return nil, ErrMissingManagementConfiguration{cluster: currentCluster}
	}

	return managementCfg, nil
}

// Purge removes the config file
func (c *Config) Purge() error {
	return os.Remove(c.loadedConfigPath)
}

// CurrentContextManifestMetadata gets manifest metadata
func (c *Config) CurrentContextManifestMetadata() (*Metadata, error) {
	manifest, err := c.CurrentContextManifest()
	if err != nil {
		return nil, err
	}
	meta := &Metadata{}
	err = util.ReadYAMLFile(manifest.MetadataPath, meta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

// DecodeAuthInfo returns authInfo with credentials decoded
func DecodeAuthInfo(authinfo *clientcmdapi.AuthInfo) (*clientcmdapi.AuthInfo, error) {
	password := authinfo.Password
	decodedPassword, err := DecodeString(password)
	if err != nil {
		return nil, ErrDecodingCredentials{Given: password}
	}
	authinfo.Password = decodedPassword

	token := authinfo.Token
	decodedToken, err := DecodeString(token)
	if err != nil {
		return nil, ErrDecodingCredentials{Given: token}
	}
	authinfo.Token = decodedToken

	clientCert := authinfo.ClientCertificate
	decodedClientCertificate, err := DecodeString(clientCert)
	if err != nil {
		return nil, ErrDecodingCredentials{Given: clientCert}
	}
	authinfo.ClientCertificate = decodedClientCertificate

	clientKey := authinfo.ClientKey
	decodedClientKey, err := DecodeString(clientKey)
	if err != nil {
		return nil, ErrDecodingCredentials{Given: clientKey}
	}
	authinfo.ClientKey = decodedClientKey
	return authinfo, nil
}

// EncodeAuthInfo returns authInfo with credentials base64 encoded
func EncodeAuthInfo(authinfo *clientcmdapi.AuthInfo) *clientcmdapi.AuthInfo {
	authinfo.Password = EncodeString(authinfo.Password)
	authinfo.Token = EncodeString(authinfo.Token)
	authinfo.ClientCertificate = EncodeString(authinfo.ClientCertificate)
	authinfo.ClientKey = EncodeString(authinfo.ClientKey)
	return authinfo
}
