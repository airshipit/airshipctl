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

	kubeconfig "k8s.io/client-go/tools/clientcmd/api"

	"opendev.org/airship/airshipctl/pkg/util"
)

// Called from root to Load the initial configuration
func (c *Config) LoadConfig(configFileArg string, kPathOptions *clientcmd.PathOptions) error {
	err := c.loadFromAirConfig(configFileArg)
	if err != nil {
		return err
	}

	// Load or initialize the kubeconfig object from a file
	err = c.loadKubeConfig(kPathOptions)
	if err != nil {
		return err
	}

	// Lets navigate through the kubeconfig to populate the references in airship config
	return c.reconcileConfig()
}

func (c *Config) loadFromAirConfig(configFileArg string) error {
	// If it exists,  Read the ConfigFile data
	// Only care about the errors here, because there is a file
	// And essentially I cannot use its data.
	// airshipctl probable should stop
	if configFileArg == "" {
		return errors.New("Configuration file location was not provided.")
	}
	// Remember where I loaded the Config from
	c.loadedConfigPath = configFileArg
	// If I have a file to read, load from it

	if _, err := os.Stat(configFileArg); os.IsNotExist(err) {
		return nil
	}
	return util.ReadYAMLFile(configFileArg, c)
}

func (c *Config) loadKubeConfig(kPathOptions *clientcmd.PathOptions) error {
	// Will need this for Persisting the changes
	c.loadedPathOptions = kPathOptions
	// Now at this point what I load might not reflect the associated kubeconfig yet
	kConfig, err := kPathOptions.GetStartingConfig()
	if err != nil {
		return err
	}
	// Store the kubeconfig object into an airship managed kubeconfig object
	c.kubeConfig = kConfig

	return nil
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
	// Specially useful if the cnofig is loaded during a get operation
	// If it was a Set this would have happened eventually any way
	if persistIt {
		return c.PersistConfig()
	}
	return nil
}

func (c *Config) reconcileClusters() (map[string]*ClusterComplexName, bool) {
	updatedClusters := make(map[string]*kubeconfig.Cluster)
	updatedClusterNames := make(map[string]*ClusterComplexName)
	persistIt := false
	for key, cluster := range c.kubeConfig.Clusters {

		clusterComplexName := NewClusterComplexName()
		clusterComplexName.FromName(key)
		// Lets check if the cluster from the kubeconfig file complies with the complex naming convention
		if !clusterComplexName.validName() {
			clusterComplexName.SetDefaultType()
			// Lets update the kubeconfig with proper airship name
			updatedClusters[clusterComplexName.Name()] = cluster

			// Remember name changes since Contexts has to be updated as well for this clusters
			updatedClusterNames[key] = clusterComplexName
			persistIt = true

			if c.kubeConfig.Clusters[key] == nil {
				c.kubeConfig.Clusters[key] = updatedClusters[key]
			}
			// Otherwise this is a cluster that didnt have an airship cluster type, however when you added the cluster type
			// Probable should just add a number _<COUNTER to it
		}

		// The cluster name is good at this point
		// Lets update the airship config file updated
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

		// Done updating
		// Lets remove anything that was updated
		if updatedClusterNames[key] != nil {
			delete(c.kubeConfig.Clusters, key)
		}
	}

	persistIt = c.rmConfigClusterStragglers(persistIt)

	return updatedClusterNames, persistIt

}

// Removes or Deletes Cluster configuration that exist in Airship Config
// and do not have any kubeconfig appropriate <clustername>_<clustertype>
// entries
func (c *Config) rmConfigClusterStragglers(persistIt bool) bool {
	rccs := persistIt
	// Checking if there is any Cluster reference in airship config that does not match
	// an actual Cluster struct in kubeconfig
	for key := range c.Clusters {
		for cType, cluster := range c.Clusters[key].ClusterTypes {
			if c.kubeConfig.Clusters[cluster.NameInKubeconf] == nil {
				// Instead of removing it , I could add a empty entry in kubeconfig as well
				// Will see what is more appropriae with use of Modules configuration
				delete(c.Clusters[key].ClusterTypes, cType)
				rccs = true
			}
		}
	}
	return rccs
}
func (c *Config) reconcileContexts(updatedClusterNames map[string]*ClusterComplexName) {
	for key, context := range c.kubeConfig.Contexts {
		// Check if the Cluster name referred to by the context
		// was updated during the cluster reconcile
		if updatedClusterNames[context.Cluster] != nil {
			context.Cluster = updatedClusterNames[context.Cluster].Name()
		}

		if c.Contexts[key] == nil {
			c.Contexts[key] = NewContext()
		}
		// Make sure the name matches
		c.Contexts[key].NameInKubeconf = context.Cluster

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
	//  - if the airship current context is valid, then updated kubeconfiug CC
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
	c.kubeConfig.CurrentContext = ""
	c.CurrentContext = ""
}

// This is called by users of the config to make sure that they have
// A complete configuration before they try to use it.
// What is a Complete configuration:
// Should be :
//	At least 1 cluster defined
//	At least 1 authinfo (user) defined
//	At least 1 context defined
//	The current context properly associated with an existsing context
// 	At least one Manifest defined
//
func (c *Config) EnsureComplete() error {
	if len(c.Clusters) == 0 {
		return errors.New("Config: At least one cluster needs to be defined")
	}
	if len(c.AuthInfos) == 0 {
		return errors.New("Config: At least one Authentication Information (User) needs to be defined")
	}

	if len(c.Contexts) == 0 {
		return errors.New("Config: At least one Context needs to be defined")
	}

	if c.CurrentContext == "" || c.Contexts[c.CurrentContext] == nil {
		return errors.New("Config: Current Context is not defined, or it doesnt identify a defined Context")
	}
	return nil
}

// This function is called to update the configuration in the file defined by the
// ConfigFile name
// It will completely overwrite the existing file,
//    If the file specified by ConfigFile exists ts updates with the contents of the Config object
//    If the file specified by ConfigFile does not exist it will create a new file.
func (c *Config) PersistConfig() error {
	// Dont care if the file exists or not, will create if needed
	// We are 100% overwriting the existsing file
	configyaml, err := c.ToYaml()
	if err != nil {
		return err
	}

	// WriteFile doesn't create the directory , create it if needed
	configDir := filepath.Dir(c.loadedConfigPath)
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return err
	}

	// Write the Airship Config file
	err = ioutil.WriteFile(c.loadedConfigPath, configyaml, 0644)
	if err != nil {
		return err
	}

	// Persist the kubeconfig file referenced
	if err := clientcmd.ModifyConfig(c.loadedPathOptions, *c.kubeConfig, true); err != nil {
		return err
	}

	return nil
}

func (c *Config) String() string {
	yaml, err := c.ToYaml()
	// This is hiding the error perhaps
	if err != nil {
		return ""
	}
	return string(yaml)
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

func (c *Config) LoadedPathOptions() *clientcmd.PathOptions {
	return c.loadedPathOptions
}
func (c *Config) SetLoadedPathOptions(po *clientcmd.PathOptions) {
	c.loadedPathOptions = po
}

func (c *Config) KubeConfig() *kubeconfig.Config {
	return c.kubeConfig
}

// This might be changed later to be generalized
func (c *Config) ClusterNames() []string {
	names := []string{}
	for k := range c.Clusters {
		names = append(names, k)
	}
	sort.Strings(names)
	return names

}

// Get A Cluster
func (c *Config) GetCluster(cName, cType string) (*Cluster, error) {
	_, exists := c.Clusters[cName]
	if !exists {
		return nil, errors.New("Cluster " + cName +
			" information was not found in the configuration.")
	}
	// Alternative to this would be enhance Cluster.String() to embedd the appropriate kubeconfig cluster information
	cluster, exists := c.Clusters[cName].ClusterTypes[cType]
	if !exists {
		return nil, errors.New("Cluster " + cName + " of type " + cType +
			" information was not found in the configuration.")
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
	kcluster := kubeconfig.NewCluster()
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
func (c *Config) GetClusters() ([]*Cluster, error) {
	clusters := []*Cluster{}
	for _, cName := range c.ClusterNames() {
		for _, ctName := range AllClusterTypes {
			cluster, err := c.GetCluster(cName, ctName)
			// Err simple means something that does not exists
			// Which is possible  since I am iterating thorugh both possible
			// cluster types
			if err == nil {
				clusters = append(clusters, cluster)
			}

		}
	}
	return clusters, nil
}

// CurrentConfig Returns the appropriate information for the current context
// Current Context holds labels for the approriate config objects
//      Cluster is the name of the cluster for this context
//      ClusterType is the name of the clustertye for this context
//      AuthInfo is the name of the authInfo for this context
//      Manifest is the default manifest to be use with this context
//      Namespace is the default namespace to use on unspecified requests
// Purpose for this method is simplifying ting the current context information
/*
func (c *Config) CurrentContext() (*Context, *Cluster, *AuthInfo, *Manifest, error) {
	if err := c.EnsureComplete(); err != nil {
		return nil, nil, nil, nil, err
	}
	currentContext := c.Contexts[c.CurrentContext]
	if currentContext == nil {
		// this should not happened
		return nil, nil, nil, nil,
			errors.New("CurrentContext was unable to find the configured current context.")
	}
	return currentContext,
		c.Clusters[currentContext.Cluster].ClusterTypes[currentContext.ClusterType],
		c.AuthInfos[currentContext.AuthInfo],
		c.Manifests[currentContext.Manifest],
		nil
}
*/

// Purge removes the config file
func (c *Config) Purge() error {
	//configFile := c.ConfigFile()
	err := os.Remove(c.loadedConfigPath)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) Equal(d *Config) bool {
	if d == nil {
		return d == c
	}
	clusterEq := reflect.DeepEqual(c.Clusters, d.Clusters)
	authInfoEq := reflect.DeepEqual(c.AuthInfos, d.AuthInfos)
	contextEq := reflect.DeepEqual(c.Contexts, d.Contexts)
	manifestEq := reflect.DeepEqual(c.Manifests, d.Manifests)
	return c.Kind == d.Kind &&
		c.APIVersion == d.APIVersion &&
		clusterEq && authInfoEq && contextEq && manifestEq &&
		c.ModulesConfig.Equal(d.ModulesConfig)
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

func (c *Cluster) KubeCluster() *kubeconfig.Cluster {
	return c.kCluster
}
func (c *Cluster) SetKubeCluster(kc *kubeconfig.Cluster) {
	c.kCluster = kc
}

// Context functions
func (c *Context) Equal(d *Context) bool {
	if d == nil {
		return d == c
	}
	return c.NameInKubeconf == d.NameInKubeconf &&
		c.Manifest == d.Manifest
}

func (c *Context) String() string {
	yaml, err := yaml.Marshal(&c)
	if err != nil {
		return ""
	}
	return string(yaml)
}

// AuthInfo functions
func (c *AuthInfo) Equal(d *AuthInfo) bool {
	if d == nil {
		return d == c
	}
	return c == d
}

func (c *AuthInfo) String() string {
	yaml, err := yaml.Marshal(&c)
	if err != nil {
		return ""
	}
	return string(yaml)
}

// Manifest functions
func (m *Manifest) Equal(n *Manifest) bool {
	if n == nil {
		return n == m
	}
	repositoryEq := reflect.DeepEqual(m.Repositories, n.Repositories)
	return repositoryEq && m.TargetPath == n.TargetPath
}

func (m *Manifest) String() string {
	yaml, err := yaml.Marshal(&m)
	if err != nil {
		return ""
	}
	return string(yaml)
}

// Repository functions
func (r *Repository) Equal(s *Repository) bool {
	if s == nil {
		return r == s
	}
	var urlMatches bool
	if r.Url != nil && s.Url != nil {
		urlMatches = (r.Url.String() == s.Url.String())
	} else {
		// this catches cases where one or both are nil
		urlMatches = (r.Url == s.Url)
	}
	return urlMatches &&
		r.Username == s.Username &&
		r.TargetPath == s.TargetPath
}
func (r *Repository) String() string {
	yaml, err := yaml.Marshal(&r)
	if err != nil {
		return ""
	}
	return string(yaml)
}

// Modules functions
func (m *Modules) Equal(n *Modules) bool {
	if n == nil {
		return n == m
	}
	return m.Dummy == n.Dummy
}
func (m *Modules) String() string {
	yaml, err := yaml.Marshal(&m)
	if err != nil {
		return ""
	}
	return string(yaml)
}

// ClusterComplexName functions
func (c *ClusterComplexName) validName() bool {
	err := ValidClusterType(c.clusterType)
	return c.clusterName != "" && err == nil
}
func (c *ClusterComplexName) FromName(clusterName string) {
	if clusterName != "" {
		userNameSplit := strings.Split(clusterName, AirshipClusterNameSep)
		if len(userNameSplit) == 2 {
			c.clusterType = userNameSplit[1]
		}
		c.clusterName = userNameSplit[0]
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
func ValidClusterType(ctype string) error {
	if ctype == Ephemeral || ctype == Target {
		return nil
	}
	return errors.New("Cluster Type must be specified. Valid values are :" + Ephemeral + " or " + Target + ".")
}

/* ______________________________
PLACEHOLDER UNTIL I IDENTIFY if CLIENTADM
HAS SOMETHING LIKE THIS
*/

func KClusterString(kCluster *kubeconfig.Cluster) string {
	yaml, err := yaml.Marshal(&kCluster)
	if err != nil {
		return ""
	}

	return string(yaml)
}
