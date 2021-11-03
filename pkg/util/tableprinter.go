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

package util

import (
	"io"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/cmd/get"
	"k8s.io/kubectl/pkg/scheme"

	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	// PhaseListFormat is used to print tables with phase list
	PhaseListFormat = "NAMESPACE:metadata.namespace,NAME:metadata.name,CLUSTER NAME:metadata.clusterName," +
		"EXECUTOR:config.executorRef.kind,DOC ENTRYPOINT:config.documentEntryPoint"
	// PlanListFormat is used to print tables with plan list
	PlanListFormat = "NAMESPACE:metadata.namespace,NAME:metadata.name,DESCRIPTION:description"
	// HostListFormat is used to print tables with host list
	HostListFormat = "NodeName:nodename,NodeID:nodeid"
)

// PrintObjects prints suitable
func PrintObjects(objects interface{}, template string, w io.Writer, noHeaders bool) error {
	tabWriter := GetNewTabWriter(w)
	defer tabWriter.Flush()

	p, err := get.NewCustomColumnsPrinterFromSpec(template,
		scheme.Codecs.UniversalDecoder(scheme.Scheme.PrioritizedVersionsAllGroups()...), noHeaders)
	if err != nil {
		return err
	}

	for _, obj := range runtimeObjectsConverter(objects) {
		if err := p.PrintObj(obj, tabWriter); err != nil {
			return err
		}
	}
	return nil
}

func runtimeObjectsConverter(object interface{}) []runtime.Object {
	var resources []runtime.Object
	if object == nil {
		return resources
	}
	value := reflect.ValueOf(object)
	switch value.Kind() {
	case reflect.Slice:
		resources = make([]runtime.Object, value.Len())
		for i := 0; i < value.Len(); i++ {
			printable, ok := value.Index(i).Interface().(runtime.Object)
			if !ok {
				log.Debugf("resource %#v is not printable, skipping", value.Index(i).Interface())
				continue
			}
			resources[i] = printable
		}
	default:
		res, ok := value.Interface().(runtime.Object)
		if !ok {
			log.Debugf("resource %#v is not printable, skipping", value.Interface())
		}
		resources = []runtime.Object{res}
	}
	return resources
}
