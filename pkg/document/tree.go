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

package document

import (
	"fmt"
	"io"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/types"

	"opendev.org/airship/airshipctl/pkg/fs"
)

// KustomNode is used to create name and data to display tree structure
type KustomNode struct {
	Name     string // name used for display purposes (cli)
	Data     string // this could be a Kustomization object, or a string containing a file path
	Children []KustomNode
	Writer   io.Writer
}

// BuildKustomTree creates a tree based on entrypoint
func BuildKustomTree(entrypoint string, writer io.Writer, manifestsDir string) (KustomNode, error) {
	fs := fs.NewDocumentFs()
	name, err := filepath.Rel(manifestsDir, entrypoint)
	if err != nil {
		name = entrypoint
	}
	root := KustomNode{
		Name:     name,
		Data:     entrypoint,
		Children: []KustomNode{},
		Writer:   writer,
	}

	resMap, err := MakeResMap(fs, entrypoint)
	if err != nil {
		return KustomNode{}, err
	}

	for sourceType, sources := range resMap {
		n := KustomNode{
			Name:   sourceType,
			Writer: writer,
		}

		for _, s := range sources {
			if !fs.IsDir(s) {
				name, err := filepath.Rel(manifestsDir, s)
				if err != nil {
					name = s
				}
				n.Children = append(n.Children, KustomNode{
					Name: name,
					Data: s,
				})
			} else {
				s = filepath.Join(s, KustomizationFile)
				child, err := BuildKustomTree(s, writer, "")
				if err != nil {
					return KustomNode{}, err
				}
				n.Children = append(n.Children, child)
			}
		}
		root.Children = append(root.Children, n)
	}
	return root, nil
}

//MakeResMap creates resmap based of kustomize types
func MakeResMap(fs fs.FileSystem, kfile string) (map[string][]string, error) {
	if fs == nil {
		return nil, fmt.Errorf("received nil filesystem")
	}
	bytes, err := fs.ReadFile(kfile)
	if err != nil {
		return nil, err
	}

	k := types.Kustomization{}
	err = k.Unmarshal(bytes)
	if err != nil {
		return nil, err
	}
	basedir := filepath.Dir(kfile)
	var resMap = make(map[string][]string)
	for _, p := range k.Resources {
		path := filepath.Join(basedir, p)
		resMap["Resources"] = append(resMap["Resources"], path)
	}

	for _, p := range k.Crds {
		path := filepath.Join(basedir, p)
		resMap["Crds"] = append(resMap["Crds"], path)
	}

	buildConfigMapAndSecretGenerator(k, basedir, resMap)

	for _, p := range k.Generators {
		path := filepath.Join(basedir, p)
		resMap["Generators"] = append(resMap["Generators"], path)
	}

	for _, p := range k.Transformers {
		path := filepath.Join(basedir, p)
		resMap["Transformers"] = append(resMap["Transformers"], path)
	}

	return resMap, nil
}

func buildConfigMapAndSecretGenerator(k types.Kustomization, basedir string, resMap map[string][]string) {
	for _, p := range k.SecretGenerator {
		for _, s := range p.FileSources {
			path := filepath.Join(basedir, s)
			resMap["SecretGenerator"] = append(resMap["SecretGenerator"], path)
		}
	}
	for _, p := range k.ConfigMapGenerator {
		for _, s := range p.FileSources {
			path := filepath.Join(basedir, s)
			resMap["ConfigMapGenerator"] = append(resMap["ConfigMapGenerator"], path)
		}
	}
}

// PrintTree prints tree view of phase
func (k KustomNode) PrintTree(prefix string) {
	if prefix == "" {
		basedir := filepath.Dir(k.Name)
		dir := filepath.Base(basedir)
		fmt.Fprintf(k.Writer, "%s [%s]\n", dir, basedir)
	}
	for i, child := range k.Children {
		var subprefix string
		knodes := GetKustomChildren(child)
		if len(knodes) > 0 {
			// we found a kustomize file, so print the subtree name first
			if i == len(k.Children)-1 {
				fmt.Fprintf(k.Writer, "%s└── %s\n", prefix, child.Name)
				subprefix = "    "
			} else {
				fmt.Fprintf(k.Writer, "%s├── %s\n", prefix, child.Name)
				subprefix = "│   "
			}
		}
		for j, c := range knodes {
			bd := filepath.Dir(c.Name)
			d := filepath.Base(bd)
			name := fmt.Sprintf("%s [%s]", d, bd)

			if j == len(knodes)-1 {
				fmt.Printf("%s%s└── %s\n", prefix, subprefix, name)
				c.PrintTree(fmt.Sprintf("%s%s    ", prefix, subprefix))
			} else {
				fmt.Printf("%s%s├── %s\n", prefix, subprefix, name)
				c.PrintTree(fmt.Sprintf("%s%s│   ", prefix, subprefix))
			}
		}
	}
}

// GetKustomChildren returns children nodes of kustomnode
func GetKustomChildren(k KustomNode) []KustomNode {
	nodes := []KustomNode{}
	for _, c := range k.Children {
		if len(c.Children) > 0 {
			nodes = append(nodes, c)
		}
	}
	return nodes
}
