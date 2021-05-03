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
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/instrumenta/kubeval/kubeval"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	schemaLocationDir = "/workdir/schemas-cache"
	fileScheme        = "file://"
	openAPISchemaFile = "openapischema"
	crdKind           = "CustomResourceDefinition"
	phaseRenderedFile = "phase-rendered.yaml"
	crdListFile       = "crd-list"
	cleanupEnv        = "VALIDATOR_PREVENT_CLEANUP"

	defaultKubernetesVersion    = "1.18.6"
	defaultStrict               = true
	defaultIgnoreMissingSchemas = false
	defaultSchemaLocation       = "https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master/"
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

// Spec holds main kubeval parameters
type Spec struct {
	// Strict disallows additional properties not in schema if set
	Strict *bool `yaml:"strict,omitempty"`

	// IgnoreMissingSchemas skips validation for resource
	// definitions without a schema.
	IgnoreMissingSchemas *bool `yaml:"ignoreMissingSchemas,omitempty"`

	// KubernetesVersion is the version of Kubernetes to validate
	// against (default "1.18.6").
	KubernetesVersion string `yaml:"kubernetesVersion,omitempty"`

	// SchemaLocation is the base URL from which to search for schemas.
	// It can be either a remote location or a local directory
	SchemaLocation string `yaml:"schemaLocation,omitempty"`

	// KindsToSkip defines Kinds which will be skipped during validation
	KindsToSkip []string `yaml:"kindsToSkip,omitempty"`
}

// CrdConfig is a small struct to process CRD list
type CrdConfig struct {
	SchemasLocation string `yaml:"schemasLocation"`
	CrdList         string `yaml:"crdList"`
}

// Filter checks each resource for validity, otherwise returning an error
func (f kubevalFilter) Filter(in []*yaml.RNode) ([]*yaml.RNode, error) {
	cfg, err := f.parseConfig()
	if err != nil {
		return nil, err
	}

	kubevalConfig := kubeval.NewDefaultConfig()
	kubevalConfig.Strict = *cfg.Strict
	kubevalConfig.IgnoreMissingSchemas = *cfg.IgnoreMissingSchemas
	kubevalConfig.KubernetesVersion = cfg.KubernetesVersion
	kubevalConfig.SchemaLocation = cfg.SchemaLocation
	kubevalConfig.AdditionalSchemaLocations = []string{fileScheme + schemaLocationDir}
	kubevalConfig.KindsToSkip = append(cfg.KindsToSkip, crdKind)

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

		// Process each additional CRD in the list (CRD -> OpenAPIV3 Schema -> Json Schema)
		if err := processCRDList(renderedCRDFile, schemasLocation); err != nil {
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

// parseConfig parses the functionConfig into a Spec struct.
func (f *kubevalFilter) parseConfig() (*Spec, error) {
	// Initialize default values
	boolPtr := func(b bool) *bool { return &b }
	cfg := &Spec{
		Strict:               boolPtr(defaultStrict),
		IgnoreMissingSchemas: boolPtr(defaultIgnoreMissingSchemas),
		KubernetesVersion:    defaultKubernetesVersion,
		SchemaLocation:       defaultSchemaLocation,
	}

	if err := yaml.Unmarshal([]byte(f.rw.FunctionConfig.MustString()), &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// processCRDList takes each CRD from crdList, converts paths to URLs,
// saves it to file and calls conversion script
func processCRDList(crdList string, schemasLocation string) error {
	configMap := CrdConfig{
		SchemasLocation: schemasLocation,
		CrdList:         crdList,
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
