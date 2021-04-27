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

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/instrumenta/kubeval/kubeval"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	manifestsMountPoint = "/manifests"
	schemaLocationDir   = "/workdir/schemas-cache"
	fileScheme          = "file"
	openAPISchemaFile   = "openapischema"
	crdKind             = "CustomResourceDefinition"
	kubevalOptsKind     = "KubevalOptions"
	phaseRenderedFile   = "phase-rendered.yaml"
	crdListFile         = "crd-list"
	cleanupEnv          = "VALIDATOR_PREVENT_CLEANUP"
	planEnv             = "VALIDATOR_PLAN_VALIDATION"

	defaultKubernetesVersion    = "1.16.0"
	defaultStrict               = true
	defaultIgnoreMissingSchemas = false
)

func main() {
	rw := &kio.ByteReadWriter{
		Reader:                os.Stdin,
		Writer:                os.Stdout,
		OmitReaderAnnotations: true,
		KeepReaderAnnotations: true,
	}
	p := kio.Pipeline{
		Inputs:  []kio.Reader{rw},                    // read the inputs into a slice
		Filters: []kio.Filter{kubevalFilter{rw: rw}}, // filters input by validation
		Outputs: []kio.Writer{rw},                    // copy the inputs to the output
	}
	if err := p.Execute(); err != nil {
		printMsg("%v\n", err)
		// Clean up working directory
		if err = os.RemoveAll(schemaLocationDir); err != nil {}
		os.Exit(1)
	}
}

// kubevalFilter implements kio.Filter
type kubevalFilter struct {
	rw *kio.ByteReadWriter
}

// CRDConfigMap is a map with appropriate validation configs for plans/phases
type CRDConfigMap map[string]*CRDSpec

// Config defines the input config schema as a struct
type Config struct {
	Spec         *Spec        `yaml:"siteConfig"`
	PlanName     string       `yaml:"planName,omitempty"`
	PlanConfigs  CRDConfigMap `yaml:"planConfigs,omitempty"`
	PhaseName    string       `yaml:"phaseName,omitempty"`
	PhaseConfigs CRDConfigMap `yaml:"phaseConfigs,omitempty"`
}

// Spec holds main kubeval parameters
type Spec struct {
	// Strict disallows additional properties not in schema if set
	Strict bool `yaml:"strict,omitempty"`

	// IgnoreMissingSchemas skips validation for resource
	// definitions without a schema.
	IgnoreMissingSchemas bool `yaml:"ignoreMissingSchemas,omitempty"`

	// KubernetesVersion is the version of Kubernetes to validate
	// against (default "master").
	KubernetesVersion string `yaml:"kubernetesVersion,omitempty"`
}

// CRDSpec holds special options for plan/phase which kinds to skip and which additional CRDs to include
type CRDSpec struct {
	// KindsToSkip defines Kinds which will be skipped during validation
	KindsToSkip []string `yaml:"kindsToSkip,omitempty"`

	// CRDList defines additional CRD locations
	CRDList []string `yaml:"crdList,omitempty"`
}

// CrdConfig is a small struct to process CRD list
type CrdConfig struct {
	SchemasLocation string   `yaml:"schemasLocation"`
	CrdList         []string `yaml:"crdList"`
}

// Filter checks each resource for validity, otherwise returning an error
func (f kubevalFilter) Filter(in []*yaml.RNode) ([]*yaml.RNode, error) {
	cfg, err := f.parseConfig()
	if err != nil {
		return nil, err
	}

	kubevalConfig := kubeval.NewDefaultConfig()
	kubevalConfig.Strict = cfg.Spec.Strict
	kubevalConfig.IgnoreMissingSchemas = cfg.Spec.IgnoreMissingSchemas
	kubevalConfig.KubernetesVersion = cfg.Spec.KubernetesVersion
	kubevalConfig.AdditionalSchemaLocations = []string{fileScheme + "://" + schemaLocationDir}
	kubevalConfig.KindsToSkip = []string{crdKind, kubevalOptsKind}
	var crdList []string

	if _, plan := os.LookupEnv(planEnv); plan {
		// Setup plan specific options
		if _, exists := cfg.PlanConfigs[cfg.PlanName]; exists {
			kubevalConfig.KindsToSkip = append(kubevalConfig.KindsToSkip, cfg.PlanConfigs[cfg.PlanName].KindsToSkip...)
			crdList = cfg.PlanConfigs[cfg.PlanName].CRDList
		}
	} else {
		// Setup phase specific options
		if _, exists := cfg.PhaseConfigs[cfg.PhaseName]; exists {
			kubevalConfig.KindsToSkip = append(kubevalConfig.KindsToSkip, cfg.PhaseConfigs[cfg.PhaseName].KindsToSkip...)
			crdList = cfg.PhaseConfigs[cfg.PhaseName].CRDList
		}
	}

	// Calculate schema location directory for kubeval and openapi2jsonschema based on options
	schemasLocation := filepath.Join(schemaLocationDir,
		fmt.Sprintf("v%s-%s", kubevalConfig.KubernetesVersion, "standalone"))
	if kubevalConfig.Strict {
		schemasLocation += "-strict"
	}
	// Create it if doesn't exist
	if _, err := os.Stat(schemasLocation); os.IsNotExist(err) {
		if err = os.MkdirAll(schemasLocation, 0755); err != nil {
			return nil, err
		}
	}

	// Filter CRDs from input
	crdRNodes, err := filterCRD(in)
	if err != nil {
		return nil, err
	}

	if len(crdRNodes) > 0 {
		// Save filtered CRDs in file to future processing
		renderedCRDFile := filepath.Join(schemaLocationDir, phaseRenderedFile)
		buf := bytes.Buffer{}
		for _, rNode := range crdRNodes {
			buf.Write([]byte("---\n" + rNode.MustString()))
		}
		if err = ioutil.WriteFile(renderedCRDFile, buf.Bytes(), 0600); err != nil {
			return nil, err
		}
		// Prepend rendered CRD to give them priority for processing
		crdList = append([]string{renderedCRDFile}, crdList...)
	}

	if len(crdList) > 0 {
		// Process each additional CRD in the list (CRD -> OpenAPIV3 Schema -> Json Schema)
		if err := processCRDList(crdList, schemasLocation); err != nil {
			return nil, err
		}
	}

	// Validate each Resource
	for _, r := range in {
		meta, err := r.GetMeta()
		if err != nil {
			return nil, err
		}

		if err := validate(r.MustString(), kubevalConfig); err != nil {
			// if there's an issue found with document - it will be printed as well
			printMsg("Resource invalid: (Kind: %s, Name: %s)\n---\n%s---\n", meta.Kind, meta.Name, r.MustString())
			return nil, err
		}
		// inform document is ok
		printMsg("Resource valid: (Kind: %s, Name: %s)\n", meta.Kind, meta.Name)
	}

	// if prevent cleanup variable is not set then cleanup working directory
	if _, cleanup := os.LookupEnv(cleanupEnv); !cleanup {
		if err := os.RemoveAll(schemaLocationDir); err != nil {
			return nil, err
		}
	}
	// Don't return output list, we satisfied with exit code and stdout/stderr
	return nil, nil
}

// filterCRD filters CRD documents from input slice of *yaml.RNodes
func filterCRD(in []*yaml.RNode) ([]*yaml.RNode, error) {
	var out []*yaml.RNode
	for _, r := range in {
		meta, err := r.GetMeta()
		if err != nil {
			return nil, err
		}
		if meta.Kind == crdKind {
			out = append(out, r)
		}
	}
	return out, nil
}

// validate runs kubeval.Validate and analyzes results
func validate(r string, config *kubeval.Config) error {
	results, err := kubeval.Validate([]byte(r), config)
	if err != nil {
		return err
	}

	return checkResults(results)
}

// checkResults processes results of validation, so we can filter some of the errors ourselves
func checkResults(results []kubeval.ValidationResult) error {
	if len(results) == 0 {
		return nil
	}

	var errs []string
	for _, r := range results {
		for _, e := range r.Errors {
			errs = append(errs, e.String())
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

// parseConfig parses the functionConfig into an Config struct.
func (f *kubevalFilter) parseConfig() (*Config, error) {
	// Initialize default values
	cfg := &Config{
		Spec: &Spec{
			Strict:               defaultStrict,
			IgnoreMissingSchemas: defaultIgnoreMissingSchemas,
			KubernetesVersion:    defaultKubernetesVersion,
		},
	}

	if err := yaml.Unmarshal([]byte(f.rw.FunctionConfig.MustString()), &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// processCRDList takes each CRD from crdList, converts paths to URLs,
// saves it to file and calls conversion script
func processCRDList(crdList []string, schemasLocation string) error {
	configMap := CrdConfig{
		SchemasLocation: schemasLocation,
		CrdList:         []string{},
	}
	// Walk through all additional CRD
	for _, crdPath := range crdList {
		// Parse provided CRD path as URL
		crdURL, err := url.Parse(crdPath)
		if err != nil {
			return err
		}
		// Add 'file' scheme if not specified and convert relative path to absolute
		if crdURL.Scheme == "" {
			crdURL.Scheme = fileScheme
			if filepath.Base(crdPath) != phaseRenderedFile {
				crdURL.Path = filepath.Join(manifestsMountPoint, crdPath)
			}
		}
		configMap.CrdList = append(configMap.CrdList, crdURL.String())
	}

	crdData, err := yaml.Marshal(configMap)
	if err != nil {
		return err
	}
	crdListFile := filepath.Join(schemaLocationDir, crdListFile)
	if err := ioutil.WriteFile(crdListFile, crdData, 0600); err != nil {
		return err
	}
	return openAPI2Json(schemasLocation)
}

// Convert OpenAPI schemas to JSON
func openAPI2Json(schemasLocation string) error {
	printMsg("Converting OpenAPI schemas to JSON\n")
	openAPISchemaPath := filepath.Join(schemaLocationDir, openAPISchemaFile)

	openAPI2JsonCmd := exec.Command("extract-openapi.py", "--strict", "--expanded", "--stand-alone",
		"--kubernetes", "-o", schemasLocation, openAPISchemaPath)
	// allows to observe realtime output from script
	w := io.Writer(os.Stderr)
	openAPI2JsonCmd.Stdout = w
	openAPI2JsonCmd.Stderr = w
	return openAPI2JsonCmd.Run()
}

// printMsg is a convenient function to print output to stderr
func printMsg(format string, a ...interface{}) {
	if _, err := fmt.Fprintf(os.Stderr, format, a...); err != nil {}
}
