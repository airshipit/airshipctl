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
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"

	sprig "github.com/Masterminds/sprig/v3"
)

const (
	beginCertificate = "-----BEGIN CERTIFICATE-----"
	endCertificate   = "-----END CERTIFICATE-----"
)

var (
	// fastCertKeyAlgos is the list of private key algorithms that are supported for certificate use, and
	// are fast to generate.
	fastCertKeyAlgos = []string{
		"ecdsa",
		"ed25519",
	}
)

// copy needed tests from https://github.com/Masterminds/sprig/blob/master/crypto_test.go
func testGenCAEx(t *testing.T, keyAlgo *string, subj, expCN string) {
	var genCAExpr string
	if keyAlgo == nil {
		genCAExpr = "genCAEx"
	} else {
		genCAExpr = fmt.Sprintf(`genPrivateKey "%s" | genCAWithKeyEx`, *keyAlgo)
	}

	tpl := fmt.Sprintf(
		`{{- $ca := %s "%s" 365 }}
{{ $ca.Cert }}
`,
		genCAExpr,
		subj,
	)
	out, err := runRaw(tpl, nil)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, out, beginCertificate)
	assert.Contains(t, out, endCertificate)

	decodedCert, _ := pem.Decode([]byte(out))
	assert.Nil(t, err)
	cert, err := x509.ParseCertificate(decodedCert.Bytes)
	assert.Nil(t, err)

	assert.Equal(t, expCN, cert.Subject.CommonName)
	assert.True(t, cert.IsCA)
}

func TestGenCAEx(t *testing.T) {
	testGenCAEx(t, nil, "foo ca", "foo ca")
	testGenCAEx(t, nil, "/CN=bar ca", "bar ca")
	for i, keyAlgo := range fastCertKeyAlgos {
		t.Run(keyAlgo, func(t *testing.T) {
			testGenCAEx(t, &fastCertKeyAlgos[i], "foo ca", "foo ca")
			testGenCAEx(t, &fastCertKeyAlgos[i], "/CN=bar ca", "bar ca")
		})
	}
}

func testGenSignedCertEx(t *testing.T, caKeyAlgo, certKeyAlgo *string, subj, expCn string) {
	const (
		ip1  = "10.0.0.1"
		ip2  = "10.0.0.2"
		dns1 = "bar.com"
		dns2 = "bat.com"
	)

	var genCAExpr, genSignedCertExpr string
	if caKeyAlgo == nil {
		genCAExpr = "genCAEx"
	} else {
		genCAExpr = fmt.Sprintf(`genPrivateKey "%s" | genCAWithKeyEx`, *caKeyAlgo)
	}
	if certKeyAlgo == nil {
		genSignedCertExpr = "genSignedCertEx"
	} else {
		genSignedCertExpr = fmt.Sprintf(`genPrivateKey "%s" | genSignedCertWithKeyEx`, *certKeyAlgo)
	}

	tpl := fmt.Sprintf(
		`{{- $ca := %s "foo" 3650 }}
{{- $cert := %s "%s" (list "%s" "%s") (list "%s" "%s") 365 $ca }}
{{ $cert.Cert }}`,
		genCAExpr,
		genSignedCertExpr,
		subj,
		ip1,
		ip2,
		dns1,
		dns2,
	)

	out, err := runRaw(tpl, nil)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, out, beginCertificate)
	assert.Contains(t, out, endCertificate)

	decodedCert, _ := pem.Decode([]byte(out))
	assert.Nil(t, err)
	cert, err := x509.ParseCertificate(decodedCert.Bytes)
	assert.Nil(t, err)

	assert.Equal(t, expCn, cert.Subject.CommonName)
	assert.Equal(t, 1, cert.SerialNumber.Sign())
	assert.Equal(t, 2, len(cert.IPAddresses))
	assert.Equal(t, ip1, cert.IPAddresses[0].String())
	assert.Equal(t, ip2, cert.IPAddresses[1].String())
	assert.Contains(t, cert.DNSNames, dns1)
	assert.Contains(t, cert.DNSNames, dns2)
	assert.False(t, cert.IsCA)
}

func TestGenSignedCertEx(t *testing.T) {
	testGenSignedCertEx(t, nil, nil, "foo ca", "foo ca")
	testGenSignedCertEx(t, nil, nil, "/CN=bar ca", "bar ca")
	for i, caKeyAlgo := range fastCertKeyAlgos {
		for j, certKeyAlgo := range fastCertKeyAlgos {
			t.Run(fmt.Sprintf("%s-%s", caKeyAlgo, certKeyAlgo), func(t *testing.T) {
				testGenSignedCertEx(t, &fastCertKeyAlgos[i], &fastCertKeyAlgos[j], "foo ca", "foo ca")
				testGenSignedCertEx(t, &fastCertKeyAlgos[i], &fastCertKeyAlgos[j], "/CN=bar ca", "bar ca")
			})
		}
	}
}

// runRaw runs a template with the given variables and returns the result.
func runRaw(tpl string, vars interface{}) (string, error) {
	funcMap := sprig.TxtFuncMap()
	for i, v := range GenericFuncMap() {
		funcMap[i] = v
	}
	t := template.Must(template.New("test").Funcs(funcMap).Parse(tpl))
	var b bytes.Buffer
	err := t.Execute(&b, vars)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
