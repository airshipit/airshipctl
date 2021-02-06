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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/bootstrap/cloudinit"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/util"
)

const (
	builderConfigFileName = "builder-conf.yaml"
	userDataFileName      = "user-data"
	networkConfigFileName = "network-data"
)

func bundleFromRNodes(rnodes []*yaml.RNode) (document.Bundle, error) {
	p := provider.NewDefaultDepProvider()
	resmapFactory := resmap.NewFactory(p.GetResourceFactory(), p.GetConflictDetectorFactory())
	resmap, err := resmapFactory.NewResMapFromRNodeSlice(rnodes)
	if err != nil {
		return &document.BundleFactory{}, err
	}
	return &document.BundleFactory{
		ResMap: resmap,
	}, nil
}

func docFromRNode(rnode *yaml.RNode) (document.Document, error) {
	rnodes := []*yaml.RNode{rnode}
	bundle, err := bundleFromRNodes(rnodes)
	if err != nil {
		return nil, err
	}
	collection, err := bundle.GetAllDocuments()
	if err != nil {
		return nil, err
	}
	if len(collection) == 0 {
		return nil, errors.New("Error while converting RNode to Document: empty document bundle")
	}
	return collection[0], nil
}

func main() {
	resourceList := &framework.ResourceList{}
	cmd := framework.Command(resourceList, func() error {
		functionConfig, ok := resourceList.FunctionConfig.(*yaml.RNode)
		if !ok {
			return errors.New("Error while type assert of FunctionConfig")
		}

		functionConfigDocument, err := docFromRNode(functionConfig)
		if err != nil {
			return err
		}
		functionConfigYaml, err := functionConfigDocument.AsYAML()
		if err != nil {
			return err
		}

		isoConfiguration := &v1alpha1.IsoConfiguration{}
		err = functionConfigDocument.ToAPIObject(isoConfiguration, v1alpha1.Scheme)
		if err != nil {
			return err
		}

		docBundle, err := bundleFromRNodes(resourceList.Items)
		if err != nil {
			return err
		}

		userData, netConf, err := cloudinit.GetCloudData(
			docBundle,
			isoConfiguration.Isogen.UserDataSelector,
			isoConfiguration.Isogen.UserDataKey,
			isoConfiguration.Isogen.NetworkConfigSelector,
			isoConfiguration.Isogen.NetworkConfigKey,
		)
		if err != nil {
			return err
		}

		functionSpec := runtimeutil.GetFunctionSpec(functionConfig)
		configPath := functionSpec.Container.StorageMounts[0].DstPath

		fls := make(map[string][]byte)
		fls[filepath.Join(configPath, userDataFileName)] = userData
		fls[filepath.Join(configPath, networkConfigFileName)] = netConf
		fls[filepath.Join(configPath, builderConfigFileName)] = functionConfigYaml

		if err = util.WriteFiles(fls, 0600); err != nil {
			return err
		}

		resourceList.Items = []*yaml.RNode{}
		return nil
	})

	if err := cmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
