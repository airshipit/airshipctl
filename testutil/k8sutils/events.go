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

package k8sutils

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/apply/event"
	applyevent "sigs.k8s.io/cli-utils/pkg/apply/event"
	clicommon "sigs.k8s.io/cli-utils/pkg/common"
	"sigs.k8s.io/cli-utils/pkg/object"

	"opendev.org/airship/airshipctl/pkg/document"
)

// SuccessEvents returns list of events that constitute a successful cli utils apply
func SuccessEvents() []applyevent.Event {
	cm := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       document.ConfigMapKind,
			APIVersion: document.ConfigMapVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", "airshipit", "inventoryID"),
			Namespace: "airshipit",
			Labels: map[string]string{
				clicommon.InventoryLabel: "inventoryID",
			},
		},
		Data: map[string]string{
			"test_test-rc__ReplicationController": "",
		},
	}
	return []applyevent.Event{
		{
			Type: applyevent.InitType,
			InitEvent: applyevent.InitEvent{
				ResourceGroups: []applyevent.ResourceGroup{
					{
						Action: applyevent.ApplyAction,
						Identifiers: []object.ObjMetadata{
							{
								Namespace: "test",
								Name:      "test-rc",
								GroupKind: schema.GroupKind{
									Group: "",
									Kind:  "ReplicationController",
								},
							},
							{
								Namespace: "airshipit",
								Name:      "airshipit-test-bundle-4bf1e4a",
								GroupKind: schema.GroupKind{
									Group: "",
									Kind:  "ConfigMap",
								},
							},
						},
					},
					{
						Action:      applyevent.PruneAction,
						Identifiers: []object.ObjMetadata{},
					},
				},
			},
		},
		{
			Type: applyevent.ApplyType,
			ApplyEvent: applyevent.ApplyEvent{
				Type:      applyevent.ApplyEventResourceUpdate,
				Operation: applyevent.Configured,
				Object:    cm,
			},
		},
		{
			Type: applyevent.ApplyType,
			ApplyEvent: applyevent.ApplyEvent{
				Type: applyevent.ApplyEventCompleted,
			},
		},
	}
}

// ErrorEvents return a list of events with error
func ErrorEvents() []applyevent.Event {
	return []applyevent.Event{
		{
			Type: event.ErrorType,
			ErrorEvent: event.ErrorEvent{
				Err: fmt.Errorf("apply-error"),
			},
		},
	}
}
