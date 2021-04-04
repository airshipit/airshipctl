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
	"os"
	"path/filepath"
	"sort"

	"sigs.k8s.io/yaml"

	"opendev.org/airship/airshipctl/pkg/fs"
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

	// CurrentContext is the name of the context that you would like to use by default
	CurrentContext string `json:"currentContext"`

	// Management configuration defines management information for all baremetal hosts in a cluster.
	ManagementConfiguration map[string]*ManagementConfiguration `json:"managementConfiguration"`

	// loadedConfigPath is the full path to the the location of the config
	// file from which this config was loaded
	// +not persisted in file
	loadedConfigPath string
	fileSystem       fs.FileSystem
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
		cfg := NewEmptyConfig()

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

// CreateConfig saves default config to the specified path
func CreateConfig(airshipConfigPath string, overwrite bool) error {
	cfg := NewConfig()
	cfg.initConfigPath(airshipConfigPath)
	return cfg.PersistConfig(overwrite)
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
	data, err := c.fileSystem.ReadFile(c.loadedConfigPath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
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

// SetFs allows to set custom filesystem used in Config object. Required for unit tests
func (c *Config) SetFs(fsys fs.FileSystem) {
	c.fileSystem = fsys
}

// PersistConfig updates the airshipctl config file to match
// the current Config object.
// If file did not previously exist, the file will be created.
// The file will be overwritten if overwrite argument set to true
func (c *Config) PersistConfig(overwrite bool) error {
	if _, err := os.Stat(c.loadedConfigPath); err == nil && !overwrite {
		return ErrConfigFileExists{Path: c.loadedConfigPath}
	}

	airshipConfigYaml, err := c.ToYaml()
	if err != nil {
		return err
	}

	// WriteFile doesn't create the directory, create it if needed
	dir := c.fileSystem.Dir(c.loadedConfigPath)
	err = c.fileSystem.MkdirAll(dir)
	if err != nil {
		return err
	}

	// Change the permission of directory
	err = c.fileSystem.Chmod(dir, os.FileMode(c.Permissions.DirectoryPermission))
	if err != nil {
		return err
	}

	// Write the Airship Config file
	err = c.fileSystem.WriteFile(c.loadedConfigPath, airshipConfigYaml)
	if err != nil {
		return err
	}

	// Change the permission of config file
	err = c.fileSystem.Chmod(c.loadedConfigPath, os.FileMode(c.Permissions.FilePermission))
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
		return nil, ErrMissingConfig{What: fmt.Sprintf("context with name '%s'", cName)}
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
func (c *Config) AddContext(ctxName string, opts ...ContextOption) *Context {
	// Create the new Airship config context
	nContext := NewContext()
	c.Contexts[ctxName] = nContext

	// Ok , I have initialized structs for the Context information
	// We can use Modify to populate the correct information
	c.ModifyContext(nContext, opts...)
	return nContext
}

// ModifyContext updates Context object with given context options
func (c *Config) ModifyContext(context *Context, opts ...ContextOption) {
	for _, o := range opts {
		o(context)
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

	manifest, exist := c.Manifests[currentContext.Manifest]
	if !exist {
		return nil, ErrMissingConfig{What: "manifest named " + currentContext.Manifest}
	}

	return manifest, nil
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
		return "", ErrMissingRepositoryName{RepoType: "phase"}
	}
	return util.GitDirNameFromURL(repo.URL()), nil
}

// CurrentContextInventoryRepositoryName returns phase inventory directory from current context's manifest
// if it is not defined PhaseRepositoryName will be used instead
func (c *Config) CurrentContextInventoryRepositoryName() (string, error) {
	ccm, err := c.CurrentContextManifest()
	if err != nil {
		return "", err
	}
	repoName := ccm.InventoryRepositoryName
	if repoName == "" {
		repoName = ccm.PhaseRepositoryName
	}
	repo, exist := ccm.Repositories[repoName]
	if !exist {
		return "", ErrMissingRepositoryName{RepoType: "inventory"}
	}
	return util.GitDirNameFromURL(repo.URL()), nil
}

// GetManifest returns a Manifest instance
func (c *Config) GetManifest(name string) (*Manifest, error) {
	manifest, exists := c.Manifests[name]
	if !exists {
		return nil, ErrMissingConfig{What: fmt.Sprintf("manifest with name '%s'", name)}
	}
	return manifest, nil
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

// CurrentContextManagementConfig returns the management options for the current context
func (c *Config) CurrentContextManagementConfig() (*ManagementConfiguration, error) {
	currentContext, err := c.GetCurrentContext()
	if err != nil {
		return nil, err
	}

	if currentContext.ManagementConfiguration == "" {
		return nil, ErrMissingConfig{
			What: fmt.Sprintf("no management config listed for context '%s'", c.CurrentContext),
		}
	}

	managementCfg, exists := c.ManagementConfiguration[currentContext.ManagementConfiguration]
	if !exists {
		return nil, ErrMissingManagementConfiguration{contextName: c.CurrentContext}
	}

	return managementCfg, nil
}

// Purge removes the config file
func (c *Config) Purge() error {
	return c.fileSystem.RemoveAll(c.loadedConfigPath)
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

	data, err := c.fileSystem.ReadFile(filepath.Join(manifest.TargetPath, phaseRepoDir, manifest.MetadataPath))
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, meta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

// WorkDir returns working directory for airshipctl. Creates if it doesn't exist
func (c *Config) WorkDir() (dir string, err error) {
	dir = filepath.Join(util.UserHomeDir(), AirshipConfigDir)
	// if not dir, create it
	if !c.fileSystem.IsDir(dir) {
		err = c.fileSystem.MkdirAll(dir)
	}
	return dir, err
}

// AddManagementConfig creates a new instance of ManagementConfig object
func (c *Config) AddManagementConfig(mgmtCfgName string, opts ...ManagementConfigOption) *ManagementConfiguration {
	// Create the new Airshipctl config ManagementConfig
	nMgmtCfg := NewManagementConfiguration()
	c.ManagementConfiguration[mgmtCfgName] = nMgmtCfg

	// We can use Modify to populate the correct information
	c.ModifyManagementConfig(nMgmtCfg, opts...)
	return nMgmtCfg
}

// ModifyManagementConfig updates ManagementConfig object with given options
func (c *Config) ModifyManagementConfig(mgmtConfig *ManagementConfiguration, opts ...ManagementConfigOption) {
	for _, o := range opts {
		o(mgmtConfig)
	}
}
