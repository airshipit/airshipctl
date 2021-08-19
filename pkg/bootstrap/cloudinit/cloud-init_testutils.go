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

package cloudinit

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/document"
	testdoc "opendev.org/airship/airshipctl/testutil/document"
)

type testData struct {
	docsCfg           []string
	secret            string
	ephUser           string
	ephNode           string
	bmh               string
	selectorNamespace string
	selectorName      string
	mockErr1          interface{}
	mockErr2          interface{}
	mockErr3          interface{}
}

const (
	ephNode     = "airshipit.org/ephemeral-node in (True, true)"
	secret      = "Secret"
	bmh         = "BareMetalHost"
	ephUser     = "airshipit.org/ephemeral-user-data in (True, true)"
	ephUserData = `apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-user-data: 'true'
  name: networkdatamalformed-malformed
type: Opaque
stringData:
  userData: cloud-init
`
	ephUserDataMalformed = `apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-user-data: 'true'
    test: userdatamalformed
  name: userdatamalformed-somesecret
type: Opaque
stringData:
  no-user-data: this secret has the right label but is missing the 'user-data' key
`
	ephUserDataNoLabel = `apiVersion: v1
kind: Secret
metadata:
  labels:
    test: userdatamissing
  name: userdatamissing-somesecret
type: Opaque
stringData:
  userData: "this secret lacks the label airshipit.org/ephemeral-user-data: true"`
	ephUserDataCredentials = `apiVersion: v1
kind: Secret
metadata:
  labels:
  name: master-1-bmc
type: Opaque
stringData:
  username: foobar
  password: goober
`
	bmhMaster1 = `apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  labels:
    airshipit.org/ephemeral-node: 'true'
  name: master-1
spec:
  bmc:
    address: ipmi://127.0.0.1
    credentialsName: master-1-bmc
  networkData:
    name: master-1-networkdata
    namespace: metal3
`
	netDataMalFormed = `apiVersion: v1
kind: Secret
metadata:
  labels:
  name: networkdatamalformed-malformed
  namespace: malformed
type: Opaque
stringData:
  no-net-data-key: the required 'net-data' key is missing
`
	bmhDuplicate = `apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  labels:
    test: ephemeralduplicate
    airshipit.org/ephemeral-node: 'true'
  name: ephemeralduplicate-master-1
`
	ephMissing = `apiVersion: v1
kind: Secret
metadata:
  labels:
    test: ephemeralmissing
  name: ephemeralmissing
type: Opaque
`
	ephNetValid = `apiVersion: v1
kind: Secret
metadata:
  labels:
    test: validdocset
  name: master-1-networkdata
  namespace: metal3
type: Opaque
stringData:
  networkData: net-config
`
)

func getValidDocSet() []string {
	return []string{
		ephUserData,
		bmhMaster1,
		ephNetValid,
		ephUserDataCredentials,
	}
}

func getEphemeralMissing() []string {
	return []string{
		ephUserData,
		ephMissing,
		bmhMaster1,
	}
}

func getEphemeralDuplicate() []string {
	return []string{
		ephUserData,
		bmhDuplicate,
		bmhDuplicate,
	}
}

func getNetworkBadPointer() []string {
	return []string{
		ephUserData,
		bmhMaster1,
		ephUserDataCredentials,
	}
}

func getNetDataMalformed() []string {
	return []string{
		ephUserData,
		bmhMaster1,
		netDataMalFormed,
		ephUserDataCredentials,
	}
}

func getUserDataMalformed() []string {
	return []string{
		ephUserDataMalformed,
	}
}

func getUserDataMissing() []string {
	return []string{
		ephUserDataNoLabel,
	}
}

func createTestBundle(t *testing.T, td testData) document.Bundle {
	bundle := &testdoc.MockBundle{}
	allDocs := make([]document.Document, len(td.docsCfg))
	returnedDocs := make(map[string]document.Document)

	for i, cfg := range td.docsCfg {
		doc, err := document.NewDocumentFromBytes([]byte(cfg))
		require.NoError(t, err)
		allDocs[i] = doc
		kind, selectorMap, name, namespace := doc.GetKind(), doc.GetLabels(), doc.GetName(), doc.GetNamespace()

		// We use data in the document to determine which document should be returned from each mock
		// function call.
		if _, ok := selectorMap[strings.TrimSuffix(td.ephUser, " in (True, true)")]; ok && kind == td.secret {
			returnedDocs["userDataDoc"] = doc
			// initialize these two key-value pairs so that we avoid
			// memory address errors. These will be overwritten.
			returnedDocs["bmhDoc"] = doc
			returnedDocs["ephOrBmhDoc"] = doc
		} else if _, ok := selectorMap[strings.TrimSuffix(td.ephNode, " in (True, true)")]; ok && kind == td.bmh {
			returnedDocs["bmhDoc"] = doc
		} else if kind == td.secret && name == td.selectorName && namespace == td.selectorNamespace {
			returnedDocs["ephOrBmhDoc"] = doc
		}
	}

	bundle.On("GetAllDocuments").Return(allDocs, nil)
	bundle.On("SelectOne", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == td.secret && selector.LabelSelector == td.ephUser
	})).
		Return(returnedDocs["userDataDoc"], td.mockErr1)
	bundle.On("SelectOne", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == td.bmh && selector.LabelSelector == td.ephNode
	})).
		Return(returnedDocs["bmhDoc"], td.mockErr2)
	bundle.On("SelectOne", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == td.secret && selector.Namespace == td.selectorNamespace && selector.Name == td.selectorName
	})).
		Return(returnedDocs["ephOrBmhDoc"], td.mockErr3)
	return bundle
}
