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
	"text/template"
)

// GenericFuncMap returns a copy of the function map
func GenericFuncMap() template.FuncMap {
	gfm := make(template.FuncMap, len(genericMap))
	for k, v := range genericMap {
		gfm[k] = v
	}
	return gfm
}

var genericMap = map[string]interface{}{
	"genCAEx":                generateCertificateAuthorityEx,
	"genCAWithKeyEx":         generateCertificateAuthorityWithPEMKeyEx,
	"genSignedCertEx":        generateSignedCertificateEx,
	"genSignedCertWithKeyEx": generateSignedCertificateWithPEMKeyEx,
	"genSSHKeyPair":          genSSHKeyPair,
	"regexGen":               regexGen,
	"toYaml":                 toYaml,
	"toUint32":               toUint32,
}
