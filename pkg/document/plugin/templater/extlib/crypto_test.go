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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"crypto/x509/pkix"
)

func TestToUint32(t *testing.T) {
	assert.Equal(t, uint32(1), toUint32(1))
	assert.Equal(t, uint32(0xffffffff), toUint32(-1))
}

func TestNameFromString(t *testing.T) {
	testCases := []struct {
		in          string
		expectedOut pkix.Name
		expectedErr string
	}{
		{
			in: `Kubernetes API`,
			expectedOut: pkix.Name{
				CommonName: `Kubernetes API`,
			},
		},
		{
			in: `/CN=Kubernetes API`,
			expectedOut: pkix.Name{
				CommonName: `Kubernetes API`,
			},
		},
		{
			in: `/CN=James \"Jim\" Smith\, III+O=example`,
			expectedOut: pkix.Name{
				CommonName: `James "Jim" Smith, III`,
				Organization: []string{
					`example`,
				},
			},
		},
		{
			in: `/CN=admin/O=system:masters`,
			expectedOut: pkix.Name{
				CommonName: `admin`,
				Organization: []string{
					`system:masters`,
				},
			},
		},
		{
			in: `/C=AU/ST=Some-State/O=Internet Widgits Pty Ltd/CN=leaf`,
			expectedOut: pkix.Name{
				CommonName: `leaf`,
				Country: []string{
					`AU`,
				},
				Province: []string{
					`Some-State`,
				},
				Organization: []string{
					`Internet Widgits Pty Ltd`,
				},
			},
		},
		{
			in: `/C=AU/ST=QLD/CN=SSLeay\/rsa test cert`,
			expectedOut: pkix.Name{
				CommonName: `SSLeay/rsa test cert`,
				Country: []string{
					`AU`,
				},
				Province: []string{
					`QLD`,
				},
			},
		},
		{
			in: `/CN=CN/SERIALNUMBER=SN` +
				`/C=C1/C=C2` +
				`/O=O1/O=O2` +
				`/OU=OU1/OU=OU2` +
				`/L=L1/L=L2` +
				`/ST=ST1/ST=ST2` +
				`/STREET=S1/STREET=S2` +
				`/POSTALCODE=PC1/POSTALCODE=PC2`,
			expectedOut: pkix.Name{
				CommonName:         `CN`,
				SerialNumber:       `SN`,
				Country:            []string{`C1`, `C2`},
				Organization:       []string{`O1`, `O2`},
				OrganizationalUnit: []string{`OU1`, `OU2`},
				Locality:           []string{`L1`, `L2`},
				Province:           []string{`ST1`, `ST2`},
				StreetAddress:      []string{`S1`, `S2`},
				PostalCode:         []string{`PC1`, `PC2`},
			},
		},
		{
			in:          `/C=AU/ST=QLD/CN=SSLeay\/rsa test cert\`,
			expectedErr: `string C=AU/ST=QLD/CN=SSLeay\/rsa test cert\ terminates incorrectly`,
		},
		{
			in:          `/C=A\U/ST=QLD/CN=SSLeay\/rsa test cert`,
			expectedErr: `string C=A\U/ST=QLD/CN=SSLeay\/rsa test cert pos 5: U shouldn't follow after '\'`,
		},
		{
			in:          `/C\N=AU/ST=QLD/CN=SSLeay\/rsa test cert`,
			expectedErr: `string C\N=AU/ST=QLD/CN=SSLeay\/rsa test cert pos 3: N shouldn't follow after '\'`,
		},
		{
			in:          `/CN=AU/ST=QLD/CN=SSLeay\/rsa <>",test cert`,
			expectedErr: `string CN=AU/ST=QLD/CN=SSLeay\/rsa <>",test cert position 29: having < without '\'`,
		},
		{
			in:          `/CN=AU=AU/ST=QLD/CN=SSLeay\/rsa test cert`,
			expectedErr: `string CN=AU=AU/ST=QLD/CN=SSLeay\/rsa test cert has extra '=' on position 6`,
		},
		{
			in:          `/CN=AU/ST=QLD/CN<>",t=SSLeay\/rsa test cert`,
			expectedErr: `string CN=AU/ST=QLD/CN<>",t=SSLeay\/rsa test cert position 16: having < without '\'`,
		},
		{
			in:          `/CN=AU/ST/CN=SSLeay\/rsa <>",test cert`,
			expectedErr: `string CN=AU/ST/CN=SSLeay\/rsa <>",test cert has separator '/', but didn't have value on position 9`,
		},
		{
			in:          `/CN=AU/CN\<=SSLeay test cert`,
			expectedErr: `unsupported property CN<`,
		},
		{
			in:          `/CN=/SP=xxx/CN=SSLeay\/rsa test cert`,
			expectedErr: `unsupported property SP`,
		},
		{
			in:          `/CN=1/CN=SSLeay\/rsa test cert`,
			expectedErr: `CN is already set`,
		},
		{
			in:          `/CN=1/SERIALNUMBER=1/SERIALNUMBER=2`,
			expectedErr: `SERIALNUMBER is already set`,
		},
	}

	for _, tc := range testCases {
		r, err := nameFromString(tc.in)
		if tc.expectedErr != "" {
			assert.EqualError(t, err, tc.expectedErr)
			continue
		}
		require.NoError(t, err)
		assert.Equal(t, tc.expectedOut, *r)
	}
}
