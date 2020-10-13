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

package kyamlutils

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/client-go/util/jsonpath"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// JSONPathFilter returns RNode under JSON path
type JSONPathFilter struct {
	Kind string `yaml:"kind,omitempty"`
	// Path is jsonpath query. See http://goessner.net/articles/JsonPath/ for
	// details.
	Path string `yaml:"path,omitempty"`
	// Mutator is a function which changes filtered node
	Mutator func([]*yaml.RNode) error
	// Create empty struct if path element does not exist
	Create bool
}

type boundaries struct {
	min, max, step int
}

// Filter returns RNode identified by JSON path
func (l JSONPathFilter) Filter(rn *yaml.RNode) (*yaml.RNode, error) {
	query, err := convertFromLegacyQuery(l.Path)
	if err != nil {
		return nil, err
	}
	jp, err := jsonpath.Parse("parser", query)
	if err != nil {
		return nil, err
	}
	if len(jp.Root.Nodes) != 1 {
		return nil, ErrBadQueryFormat{Msg: "query must contain one expression"}
	}

	fieldList, ok := jp.Root.Nodes[0].(*jsonpath.ListNode)
	if !ok {
		return nil, ErrBadQueryFormat{Msg: "failed to convert root node to ListNode type"}
	}
	// Query result can be a list of values (e.g filter queries) since
	// walk is recursive we should pass list of RNode even for initial step
	rns, err := l.walk([]*yaml.RNode{rn}, fieldList)
	if len(rns) == 0 || err != nil {
		return nil, err
	}

	if l.Mutator != nil {
		if err := l.Mutator(rns); err != nil {
			return nil, err
		}
	}

	// Return first element if filtered list contins only one RNode
	// otherwise create a SequenceNode
	if len(rns) == 1 {
		return rns[0], nil
	}
	nodes := make([]*yaml.Node, len(rns))
	for i := range rns {
		nodes[i] = rns[i].YNode()
	}
	return yaml.NewRNode(&yaml.Node{Kind: yaml.SequenceNode, Content: nodes}), nil
}

func (l JSONPathFilter) walk(nodes []*yaml.RNode, fieldList *jsonpath.ListNode) (res []*yaml.RNode, err error) {
	res = nodes
	for _, field := range fieldList.Nodes {
		res, err = l.getByField(res, field)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (l JSONPathFilter) getByField(rns []*yaml.RNode, field jsonpath.Node) ([]*yaml.RNode, error) {
	switch fieldT := field.(type) {
	// Can be encountered in filter expressions e.g. [?(10 == 10)]; 10 is IntNode
	case *jsonpath.IntNode:
		return []*yaml.RNode{
			yaml.NewRNode(&yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!int",
				Value: strconv.Itoa(fieldT.Value),
			}),
		}, nil
	// Same as IntNode
	case *jsonpath.FloatNode:
		return []*yaml.RNode{
			yaml.NewRNode(&yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!float",
				Value: strconv.FormatFloat(fieldT.Value, 'f', -1, 64),
			}),
		}, nil
	// Same as IntNode
	case *jsonpath.BoolNode:
		return []*yaml.RNode{yaml.NewScalarRNode(strconv.FormatBool(fieldT.Value))}, nil
	// Same as IntNode
	case *jsonpath.TextNode:
		return []*yaml.RNode{yaml.NewScalarRNode(fieldT.Text)}, nil
	// For example spec.containers.metadata 'spec', 'containers' and
	// 'metadata' have FieldNode type
	case *jsonpath.FieldNode:
		res, err := l.doField(rns, fieldT)
		// Create field node if needed
		if len(res) == 0 && l.Create {
			res = []*yaml.RNode{yaml.NewMapRNode(nil)}
			return res, l.updateMapRNodes(rns, fieldT.Value, res[0])
		}
		return res, err
	// Field is accessing to array via direct element number or by getting
	// a slice e.g. [20] or [2:10:3]
	case *jsonpath.ArrayNode:
		return l.doArray(rns, fieldT)
	// Filter some elements of the array e.g. [?(.name == 'someName')] will
	// return all array elements which contain name field equal to 'someName'
	case *jsonpath.FilterNode:
		return l.doFilter(rns, fieldT)
	// Return all elements e.g. spec.containers[*]
	case *jsonpath.WildcardNode:
		return l.doWildcard(rns)
	}
	return nil, ErrBadQueryFormat{Msg: fmt.Sprintf("unsupported jsonpath expression %s", l.Path)}
}

func (l JSONPathFilter) doField(rns []*yaml.RNode, field *jsonpath.FieldNode) ([]*yaml.RNode, error) {
	// NOTE this code block was adapted from k8s.io/client-go/util/jsonpath
	// see evalField from jsonpath.go
	result := []*yaml.RNode{}
	if len(rns) == 0 {
		return result, nil
	}
	if field.Value == "" {
		return rns, nil
	}
	for _, rn := range rns {
		res, err := rn.Pipe(yaml.Get(field.Value))
		if err != nil {
			return nil, err
		}
		if res != nil {
			res.AppendToFieldPath(append(rn.FieldPath(), field.Value)...)
			result = append(result, res)
		}
	}
	return result, nil
}

func (l JSONPathFilter) doArray(rns []*yaml.RNode, field *jsonpath.ArrayNode) ([]*yaml.RNode, error) {
	// NOTE this code block was adapted from k8s.io/client-go/util/jsonpath
	// see evalArray from jsonpath.go
	result := []*yaml.RNode{}
	for _, rn := range rns {
		if err := yaml.ErrorIfInvalid(rn, yaml.SequenceNode); err != nil {
			return nil, err
		}
		bounds, err := l.defineArrayBoundaries(field.Params, len(rn.Content()))
		if err != nil {
			return nil, err
		}

		if bounds.min == bounds.max {
			return result, nil
		}

		for i := bounds.min; i < bounds.max; i += bounds.step {
			result = append(result, yaml.NewRNode(rn.Content()[i]))
		}
	}
	return result, nil
}

func (l JSONPathFilter) defineArrayBoundaries(params [3]jsonpath.ParamsEntry, maxLen int) (boundaries, error) {
	res := boundaries{
		min:  0,
		max:  maxLen,
		step: 1,
	}
	if params[0].Known {
		res.min = params[0].Value
	}
	if params[0].Value < 0 {
		res.min = params[0].Value + maxLen
	}
	if params[1].Known {
		res.max = params[1].Value
	}

	if params[1].Value < 0 || (params[1].Value == 0 && params[1].Derived) {
		res.max = params[1].Value + maxLen
	}

	if res.min >= maxLen || res.min < 0 {
		return boundaries{min: -1, max: -1, step: -1}, ErrIndexOutOfBound{Index: params[0].Value, Length: maxLen}
	}
	if res.max > maxLen || res.max < 0 {
		return boundaries{min: -1, max: -1, step: -1}, ErrIndexOutOfBound{Index: params[1].Value - 1, Length: maxLen}
	}
	if res.min > res.max {
		return boundaries{min: -1, max: -1, step: -1}, ErrLookup{
			fmt.Sprintf("starting index %d is greater than ending index %d",
				params[0].Value, params[1].Value),
		}
	}

	if params[2].Known {
		if params[2].Value <= 0 {
			return boundaries{min: -1, max: -1, step: -1}, ErrLookup{fmt.Sprintf("step must be > 0")}
		}
		res.step = params[2].Value
	}

	return res, nil
}

func (l JSONPathFilter) doFilter(rns []*yaml.RNode, field *jsonpath.FilterNode) ([]*yaml.RNode, error) {
	// NOTE this cond block was adapted from k8s.io/client-go/util/jsonpath
	// see evalFilter from from jsonpath.go
	result := []*yaml.RNode{}
	for _, rn := range rns {
		// Elements() will return an error if rn is not SequenceNode
		// therefore no need to check it explicitly
		items, err := rn.Elements()
		if err != nil {
			return nil, err
		}
		var left, right []*yaml.RNode
		for _, item := range items {
			pass := false
			// Evaluate left part of the filter expression against each
			// array item
			left, err = l.walk([]*yaml.RNode{item}, field.Left)
			if err != nil {
				return nil, err
			}

			if len(left) == 0 {
				continue
			}
			if len(left) > 1 {
				return nil, ErrBadQueryFormat{Msg: "can only compare one element at a time"}
			}

			// Evaluate right part of the filter expression against each
			// array item
			right, err = l.walk([]*yaml.RNode{item}, field.Right)
			if err != nil {
				return nil, err
			}

			if len(right) == 0 {
				continue
			}
			if len(right) > 1 {
				return nil, ErrBadQueryFormat{Msg: "can only compare one element at a time"}
			}

			pass, err = l.compare(field.Operator, left[0], right[0])
			if err != nil {
				return nil, err
			}
			if pass {
				result = append(result, item)
			}
		}
	}
	return result, nil
}

func (l JSONPathFilter) compare(op string, left, right *yaml.RNode) (bool, error) {
	err := yaml.ErrorIfAnyInvalidAndNonNull(yaml.ScalarNode, left, right)
	if op != "==" && err != nil {
		return false, err
	}
	switch op {
	case "<":
		return lt(left, right)
	case ">":
		return gt(left, right)
	case "==":
		return eq(left, right)
	case "!=":
		pass, err := eq(left, right)
		if err != nil {
			return false, err
		}
		return !pass, nil
	case "<=":
		return le(left, right)
	case ">=":
		return ge(left, right)
	}
	return false, ErrLookup{Msg: fmt.Sprintf("unrecognized filter operator %s", op)}
}

func (l JSONPathFilter) doWildcard(rns []*yaml.RNode) ([]*yaml.RNode, error) {
	// NOTE this code block was adapted from k8s.io/client-go/util/jsonpath
	// see evalArray from jsonpath.go
	result := []*yaml.RNode{}
	for _, rn := range rns {
		switch rn.YNode().Kind {
		case yaml.SequenceNode:
			items, err := rn.Elements()
			if err != nil {
				return nil, err
			}
			result = append(result, items...)
		case yaml.MappingNode:
			keys, err := rn.Fields()
			if err != nil {
				return nil, err
			}
			for _, key := range keys {
				result = append(result, rn.Field(key).Value)
			}
		}
	}
	return result, nil
}

func (l JSONPathFilter) updateMapRNodes(rns []*yaml.RNode, name string, node *yaml.RNode) error {
	for _, rn := range rns {
		_, err := rn.Pipe(yaml.SetField(name, node))
		if err != nil {
			return err
		}
	}
	return nil
}

func eq(a, b *yaml.RNode) (bool, error) {
	return deepEq(a.YNode(), b.YNode())
}

func lt(a, b *yaml.RNode) (bool, error) {
	if a.YNode().Tag == "!!float" || b.YNode().Tag == "!!float" {
		af, err := strconv.ParseFloat(a.YNode().Value, 64)
		if err != nil {
			return false, err
		}
		bf, err := strconv.ParseFloat(b.YNode().Value, 64)
		if err != nil {
			return false, err
		}
		return af < bf, nil
	}

	if a.YNode().Tag == "!!int" || b.YNode().Tag == "!!int" {
		ai, err := strconv.Atoi(a.YNode().Value)
		if err != nil {
			return false, err
		}
		bi, err := strconv.Atoi(b.YNode().Value)
		if err != nil {
			return false, err
		}
		return ai < bi, nil
	}
	return a.YNode().Value < b.YNode().Value, nil
}

func le(a, b *yaml.RNode) (bool, error) {
	if less, err := lt(a, b); less || err != nil {
		return less, err
	}
	return eq(a, b)
}

func gt(a, b *yaml.RNode) (bool, error) {
	return lt(b, a)
}

func ge(a, b *yaml.RNode) (bool, error) {
	return le(b, a)
}

func scalarEq(a, b *yaml.Node) (bool, error) {
	if a.Tag == "!!float" || b.Tag == "!!float" {
		af, err := strconv.ParseFloat(a.Value, 64)
		if err != nil {
			return false, err
		}
		bf, err := strconv.ParseFloat(b.Value, 64)
		if err != nil {
			return false, err
		}
		return af == bf, nil
	}

	if a.Tag == "!!int" || b.Tag == "!!int" {
		ai, err := strconv.Atoi(a.Value)
		if err != nil {
			return false, err
		}
		bi, err := strconv.Atoi(b.Value)
		if err != nil {
			return false, err
		}
		return ai == bi, nil
	}
	return a.Value == b.Value, nil
}

func deepEq(a, b *yaml.Node) (bool, error) {
	if a == nil || b == nil {
		return a == b, nil
	}
	if a.Kind == yaml.ScalarNode || b.Kind == yaml.ScalarNode {
		return scalarEq(a, b)
	}
	if len(a.Content) != len(b.Content) {
		return false, nil
	}
	for i, item := range a.Content {
		pass, err := deepEq(item, b.Content[i])
		if !pass || err != nil {
			return false, err
		}
	}
	return true, nil
}

func convertFromLegacyQuery(query string) (string, error) {
	path := strings.TrimSpace(query)
	if strings.HasPrefix(query, "{") {
		return path, nil
	}
	var result []string
	// NOTE this code is borrowed from k8sdeps module of kustomize api
	start := 0
	insideParentheses := false
	for i, c := range path {
		switch c {
		case '.':
			if !insideParentheses {
				result = convertField(path[start:i], result)
				start = i + 1
			}
		case '[':
			if !insideParentheses {
				if path[start:i] != "" {
					result = append(result, path[start:i])
				}
				start = i + 1
				insideParentheses = true
			} else {
				return "", ErrQueryConversion{Msg: "nested parentheses are not allowed", Query: path}
			}
		case ']':
			if insideParentheses {
				// Assign this index to the current
				// PathSection, save it to the result, then begin
				// a new PathSection
				result = convertFiltering(path[start:i], result)

				start = i + 1
				insideParentheses = false
			} else {
				return "", ErrQueryConversion{Msg: "invalid field path", Query: path}
			}
		}
	}
	if start < len(path) {
		result = convertField(path[start:], result)
	}
	return "{." + strings.Join(result, ".") + "}", nil
}

func convertFiltering(query string, path []string) []string {
	_, err := strconv.Atoi(query)
	result := path
	switch {
	case err == nil:
		// We have detected an integer so an array.
		result[len(result)-1] = result[len(result)-1] + "[" + query + "]"
	case strings.Contains(query, "="):
		kvFilterParts := strings.Split(query, "=")
		q := fmt.Sprintf("[?(.%s == '%s')]",
			strings.ReplaceAll(kvFilterParts[0], ".", "\\."), kvFilterParts[1])
		result[len(result)-1] += q
	default:
		// e.g. spec['containers']
		key := strings.Trim(query, "'")
		key = strings.Trim(key, "\"")
		key = strings.ReplaceAll(key, ".", "\\.")
		result = append(result, key)
	}
	return result
}

func convertField(query string, path []string) []string {
	if query == "" {
		return path
	}
	_, err := strconv.Atoi(query)
	if err != nil {
		return append(path, query)
	}

	if len(path) == 0 {
		return []string{"[" + query + "]"}
	}

	result := append([]string(nil), path...)
	result[len(result)-1] += "[" + query + "]"
	return result
}
