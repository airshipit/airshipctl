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

package certificates

import "fmt"

// ErrInvalidType gets called when the input crt/key has wrong tags
type ErrInvalidType struct {
	inputType string
}

func (e ErrInvalidType) Error() string {
	return fmt.Sprintf("Invalid certificate type %s", e.inputType)
}

// ErrMalformedCertificateData is returned when invalid crt or key is found
type ErrMalformedCertificateData struct {
	errMsg string
}

func (e ErrMalformedCertificateData) Error() string {
	return fmt.Sprintf("Malformed data; unable to extract; %s", e.errMsg)
}
