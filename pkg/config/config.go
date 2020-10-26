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
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"sigs.k8s.io/yaml"

	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/util"
)

// Where possible, json tags match the cli argument names.
// Top level config objects and all values required for proper functioning are not "omitempty".
// Any truly optional piece of config is allowed to be omitted.

// Config holds the information required by airshipctl commands
// It is somewhat a superset of what a kubeconfig looks like
type Config struct {
	// +optional
	Kind string `json:"kind,omitempty"`

	// +optional
	APIVersion string `json:"apiVersion,omitempty"`

	// Permissions is a struct of permissions for file and directory
	Permissions Permissions `json:"permissions,omitempty"`

	// Contexts is a map of referenceable names to context configs
	Contexts map[string]*Context `json:"contexts"`

	// Manifests is a map of referenceable names to documents
	Manifests map[string]*Manifest `json:"manifests"`

	// EncryptionConfigs is a map of referenceable names to encryption configs
	EncryptionConfigs map[string]*EncryptionConfig `json:"encryptionConfigs"`

	// CurrentContext is the name of the context that you would like to use by default
	CurrentContext string `json:"currentContext"`

	// Management configuration defines management information for all baremetal hosts in a cluster.
	ManagementConfiguration map[string]*ManagementConfiguration `json:"managementConfiguration"`

	// loadedConfigPath is the full path to the the location of the config
	// file from which this config was loaded
	// +not persisted in file
	loadedConfigPath string
}

// Permissions has the permissions for file and directory
type Permissions struct {
	DirectoryPermission uint32
	FilePermission      uint32
}

// Factory is a function which returns ready to use config object and error (if any)
type Factory func() (*Config, error)

// CreateFactory returns function which creates ready to use Config object
func CreateFactory(airshipConfigPath *string) Factory {
	return func() (*Config, error) {
		cfg := NewConfig()

		var acp string
		if airshipConfigPath != nil {
			acp = *airshipConfigPath
		}

		cfg.initConfigPath(acp)
		err := cfg.LoadConfig()
		if err != nil {
			// Should stop airshipctl
			log.Fatal("Failed to load or initialize config: ", err)
		}

		return cfg, cfg.EnsureComplete()
	}
}

// CreateConfig saves default config to specified paths
func CreateConfig(airshipConfigPath string) error {
	cfg := NewConfig()
	cfg.initConfigPath(airshipConfigPath)
	return cfg.PersistConfig()
}

// initConfigPath - Initializes loadedConfigPath variable for Config object
func (c *Config) initConfigPath(airshipConfigPath string) {
	switch {
	case airshipConfigPath != "":
		// The loadedConfigPath may already have been received as a command line argument
		c.loadedConfigPath = airshipConfigPath
	case os.Getenv(AirshipConfigEnv) != "":
		// Otherwise, we can check if we got the path via ENVIRONMENT variable
		c.loadedConfigPath = os.Getenv(AirshipConfigEnv)
	default:
		// Otherwise, we'll try putting it in the home directory
		c.loadedConfigPath = filepath.Join(util.UserHomeDir(), AirshipConfigDir, AirshipConfig)
	}
}

// LoadConfig populates the Config from the file found at airshipConfigPath.
// If there is no file at airshipConfigPath, this function does nothing.
// An error is returned if:
// * airshipConfigPath is the empty string
// * the file at airshipConfigPath is inaccessible
// * the file at airshipConfigPath cannot be marshaled into Config
func (c *Config) LoadConfig() error {
	// If I can read from the file, load from it
	// throw an error otherwise
	if _, err := os.Stat(c.loadedConfigPath); err != nil {
		return err
	}

	return util.ReadYAMLFile(c.loadedConfigPath, c)
}

// EnsureComplete verifies that a Config object is ready to use.
// A complete Config object meets the following criteria:
//   * At least 1 Context is defined
//   * At least 1 Manifest is defined
//   * The CurrentContext is set
//   * The CurrentContext identifies an existing Context
//   * The CurrentContext identifies an existing Manifest
func (c *Config) EnsureComplete() error {
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

// PersistConfig updates the airshipctl config file to match
// the current Config object.
// If file did not previously exist, the file will be created.
// Otherwise, the file will be overwritten
func (c *Config) PersistConfig() error {
	airshipConfigYaml, err := c.ToYaml()
	if err != nil {
		return err
	}

	// WriteFile doesn't create the directory, create it if needed
	configDir := filepath.Dir(c.loadedConfigPath)
	err = os.MkdirAll(configDir, os.FileMode(c.Permissions.DirectoryPermission))
	if err != nil {
		return err
	}

	// Write the Airship Config file
	err = ioutil.WriteFile(c.loadedConfigPath, airshipConfigYaml, os.FileMode(c.Permissions.FilePermission))
	if err != nil {
		return err
	}

	// Change the permission of directory
	err = os.Chmod(configDir, os.FileMode(c.Permissions.DirectoryPermission))
	if err != nil {
		return err
	}

	// Change the permission of config file
	err = os.Chmod(c.loadedConfigPath, os.FileMode(c.Permissions.FilePermission))
	if err != nil {
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
	nContext.NameInKubeconf = theContext.Name

	// Ok , I have initialized structs for the Context information
	// We can use Modify to populate the correct information
	c.ModifyContext(nContext, theContext)
	return nContext
}

// ModifyContext updates Context object with given given context options
func (c *Config) ModifyContext(context *Context, theContext *ContextOptions) {
	if theContext.ManagementConfiguration != "" {
		context.ManagementConfiguration = theContext.ManagementConfiguration
	}
	if theContext.Manifest != "" {
		context.Manifest = theContext.Manifest
	}
	if theContext.EncryptionConfig != "" {
		context.EncryptionConfig = theContext.EncryptionConfig
	}
}

// GetCurrentContext methods Returns the appropriate information for the current context
// Current Context holds labels for the appropriate config objects
func (c *Config) GetCurrentContext() (*Context, error) {
	currentContext, err := c.GetContext(c.CurrentContext)
	if err != nil {
		// this should not happen since Ensure Complete checks for this
		return nil, err
	}
	return currentContext, nil
}

// CurrentContextManifest returns the manifest for the current context
func (c *Config) CurrentContextManifest() (*Manifest, error) {
	currentContext, err := c.GetCurrentContext()
	if err != nil {
		return nil, err
	}

	return c.Manifests[currentContext.Manifest], nil
}

// CurrentContextTargetPath returns target path from current context's manifest
func (c *Config) CurrentContextTargetPath() (string, error) {
	ccm, err := c.CurrentContextManifest()
	if err != nil {
		return "", err
	}
	return ccm.TargetPath, nil
}

// CurrentContextPhaseRepositoryDir returns phase repository directory from current context's manifest
// E.g. let repository url be "http://dummy.org/phaserepo.git" then repo directory under targetPath is "phaserepo"
func (c *Config) CurrentContextPhaseRepositoryDir() (string, error) {
	ccm, err := c.CurrentContextManifest()
	if err != nil {
		return "", err
	}
	repo, exist := ccm.Repositories[ccm.PhaseRepositoryName]
	if !exist {
		return "", ErrMissingRepositoryName{}
	}
	return util.GitDirNameFromURL(repo.URL()), nil
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
	if theManifest.IsPhase {
		manifest.PhaseRepositoryName = theManifest.RepoName
	}
	if theManifest.SubPath != "" {
		manifest.SubPath = theManifest.SubPath
	}
	if theManifest.TargetPath != "" {
		manifest.TargetPath = theManifest.TargetPath
	}
	if theManifest.MetadataPath != "" {
		manifest.MetadataPath = theManifest.MetadataPath
	}
	// There is no repository details to be updated
	if theManifest.RepoName == "" {
		return nil
	}
	//when setting an existing repository as phase, verify whether the repository exists
	//and user is also not passing any repository URL
	if theManifest.IsPhase && theManifest.URL == "" && (manifest.Repositories[theManifest.RepoName] == nil) {
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

// GetEncryptionConfigs returns all the encryption configs associated with the Config sorted by name
func (c *Config) GetEncryptionConfigs() []*EncryptionConfig {
	keys := make([]string, 0, len(c.EncryptionConfigs))
	for name := range c.EncryptionConfigs {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	encryptionConfigs := make([]*EncryptionConfig, 0, len(c.EncryptionConfigs))
	for _, name := range keys {
		encryptionConfigs = append(encryptionConfigs, c.EncryptionConfigs[name])
	}
	return encryptionConfigs
}

// GetEncryptionConfig returns encryption configs associated with name
// Returns error if no encryption config with name is found
func (c *Config) GetEncryptionConfig(name string) (*EncryptionConfig, error) {
	encryptionConfig, exists := c.EncryptionConfigs[name]
	if !exists {
		return nil, ErrEncryptionConfigurationNotFound{
			Name: fmt.Sprintf("Encryption Config with name '%s'", name),
		}
	}
	return encryptionConfig, nil
}

// AddEncryptionConfig creates a new encryption config
func (c *Config) AddEncryptionConfig(options *EncryptionConfigOptions) *EncryptionConfig {
	encryptionConfig := &EncryptionConfig{
		EncryptionKeyFileSource: EncryptionKeyFileSource{
			EncryptionKeyPath: options.EncryptionKeyPath,
			DecryptionKeyPath: options.DecryptionKeyPath,
		},
		EncryptionKeySecretSource: EncryptionKeySecretSource{
			KeySecretName:      options.KeySecretName,
			KeySecretNamespace: options.KeySecretNamespace,
		},
	}
	if c.EncryptionConfigs == nil {
		c.EncryptionConfigs = make(map[string]*EncryptionConfig)
	}
	c.EncryptionConfigs[options.Name] = encryptionConfig
	return encryptionConfig
}

// ModifyEncryptionConfig sets existing values to existing encryption config
func (c *Config) ModifyEncryptionConfig(encryptionConfig *EncryptionConfig, options *EncryptionConfigOptions) {
	if options.EncryptionKeyPath != "" {
		encryptionConfig.EncryptionKeyPath = options.EncryptionKeyPath
	}
	if options.DecryptionKeyPath != "" {
		encryptionConfig.DecryptionKeyPath = options.DecryptionKeyPath
	}
	if options.KeySecretName != "" {
		encryptionConfig.KeySecretName = options.KeySecretName
	}
	if options.KeySecretNamespace != "" {
		encryptionConfig.KeySecretNamespace = options.KeySecretNamespace
	}
	return
}

// CurrentContextManagementConfig returns the management options for the current context
func (c *Config) CurrentContextManagementConfig() (*ManagementConfiguration, error) {
	currentContext, err := c.GetCurrentContext()
	if err != nil {
		return nil, err
	}

	if currentContext.ManagementConfiguration == "" {
		return nil, ErrMissingConfig{
			What: fmt.Sprintf("No management config listed for cluster %s", currentContext.NameInKubeconf),
		}
	}

	managementCfg, exists := c.ManagementConfiguration[currentContext.ManagementConfiguration]
	if !exists {
		return nil, ErrMissingManagementConfiguration{context: currentContext}
	}

	return managementCfg, nil
}

// CurrentContextEncryptionConfig returns the encryption config for the current context
func (c *Config) CurrentContextEncryptionConfig() (*EncryptionConfig, error) {
	currentContext, err := c.GetCurrentContext()
	if err != nil {
		return nil, err
	}

	return c.GetEncryptionConfig(currentContext.EncryptionConfig)
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
	phaseRepoDir, err := c.CurrentContextPhaseRepositoryDir()
	if err != nil {
		return nil, err
	}
	meta := &Metadata{
		// Populate with empty values to avoid nil pointers
		Inventory: &InventoryMeta{},
		PhaseMeta: &PhaseMeta{},
	}
	err = util.ReadYAMLFile(filepath.Join(manifest.TargetPath, phaseRepoDir, manifest.MetadataPath), meta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}
