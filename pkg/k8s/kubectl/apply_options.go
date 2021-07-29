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

package kubectl

import (
	"fmt"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/kubectl/pkg/cmd/apply"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// ApplyOptions is a abstraction layer
// to ApplyOptions of kubectl/apply package
type ApplyOptions struct {
	ApplyOptions *apply.ApplyOptions
}

// SetDryRun enables/disables the dry run flag in kubectl apply options
func (ao *ApplyOptions) SetDryRun(dryRun bool) {
	if dryRun {
		// --dry-run is deprecated and can be replaced with --dry-run=client.
		ao.ApplyOptions.DryRunStrategy = cmdutil.DryRunClient
	}
}

// SetPrune enables/disables the prune flag in kubectl apply options
func (ao *ApplyOptions) SetPrune(label string) {
	if label != "" {
		ao.ApplyOptions.Prune = true
		ao.ApplyOptions.Selector = label
	} else {
		ao.ApplyOptions.Prune = false
	}
}

// SetSourceFiles sets files to read for kubectl apply command
func (ao *ApplyOptions) SetSourceFiles(fileNames []string) {
	ao.ApplyOptions.DeleteOptions.Filenames = fileNames
}

// Run executes the `apply` command.
func (ao *ApplyOptions) Run() error {
	return ao.ApplyOptions.Run()
}

// NewApplyOptions is a helper function that Creates ApplyOptions of kubectl apply module
// Values set here, are default, and do not conflict with each other, can be used if you
// need `kubectl apply` functionality without calling executing command in shell
// To function properly, you may need to specify files from where to read the resources:
// SetSourceFiles of returned object has to be used for that
func NewApplyOptions(f cmdutil.Factory, streams genericclioptions.IOStreams) (*ApplyOptions, error) {
	o := apply.NewApplyOptions(streams)
	o.ServerSideApply = false
	o.ForceConflicts = false

	o.ToPrinter = func(operation string) (printers.ResourcePrinter, error) {
		o.PrintFlags.NamePrintFlags.Operation = operation
		if o.DryRunStrategy == cmdutil.DryRunClient {
			err := o.PrintFlags.Complete("%s (dry run)")
			if err != nil {
				return nil, err
			}
		}
		if o.DryRunStrategy == cmdutil.DryRunServer {
			err := o.PrintFlags.Complete("%s (server dry run)")
			if err != nil {
				return nil, err
			}
		}
		return o.PrintFlags.ToPrinter()
	}

	var err error
	o.Recorder, err = o.RecordFlags.ToRecorder()
	if err != nil {
		return nil, err
	}

	dynamicClient, err := f.DynamicClient()
	if err != nil {
		return nil, err
	}

	o.DeleteOptions, err = o.DeleteFlags.ToOptions(dynamicClient, o.IOStreams)
	if err != nil {
		return nil, err
	}

	// This can only fail if ToDiscoverClient() function fails
	o.OpenAPISchema, err = f.OpenAPISchema()
	if err != nil {
		return nil, err
	}

	o.Validator, err = f.Validator(false)
	if err != nil {
		return nil, err
	}

	o.Builder = f.NewBuilder()
	o.Mapper, err = f.ToRESTMapper()
	if err != nil {
		return nil, err
	}

	o.DynamicClient = dynamicClient

	cl := f.ToRawKubeConfigLoader()
	if cl == nil {
		return nil, fmt.Errorf("ToRawKubeConfigLoader() returned nil")
	}

	o.Namespace, o.EnforceNamespace, err = cl.Namespace()
	if err != nil {
		return nil, err
	}
	return &ApplyOptions{ApplyOptions: o}, nil
}
