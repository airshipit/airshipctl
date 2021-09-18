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

package extlib

import (
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/merge2"
)

func yOneFilter(input *yaml.RNode, cfg string) *yaml.RNode {
	flt := yFilter(cfg)
	if flt == nil {
		return nil
	}

	return yPipe(input, []interface{}{flt})
}

func kOneFilter(input []*yaml.RNode, cfg string) []*yaml.RNode {
	flt := kFilter(cfg)
	if flt == nil {
		return nil
	}

	return kPipe(input, []interface{}{flt})
}

func yMerge(src, dest *yaml.RNode) *yaml.RNode {
	res, err := merge2.Merge(
		src, dest,
		yaml.MergeOptions{
			ListIncreaseDirection: yaml.MergeOptionsListPrepend,
		})
	if err != nil {
		return nil
	}
	return res
}

func yListAppend(src, el *yaml.RNode) *yaml.RNode {
	flt := &yaml.ElementAppender{
		Elements: []*yaml.Node{
			el.YNode(),
		},
	}
	out, err := src.Pipe(flt)
	if err != nil {
		return nil
	}
	return out
}

func strToY(in string) *yaml.RNode {
	res, err := yaml.Parse(in)
	if err != nil {
		return nil
	}
	return res
}
