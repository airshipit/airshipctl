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

package certificates_test

import (
	"crypto/x509/pkix"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/util/certificates"
)

// doValidation implements test logic for Validating the created crt/key
func doValidation(t *testing.T, data string, cType certificates.CertificateType) {
	t.Helper()
	receivedType, err := certificates.ValidateRawCertificate([]byte(data))
	require.NoError(t, err)
	assert.Equal(t, cType, receivedType)
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name        string
		expectedErr string
		options     certificates.KeyPairOptions
	}{
		{
			name: "valid-private-key",
			options: certificates.KeyPairOptions{
				Subj: certificates.Subject{
					Name: pkix.Name{
						CommonName: "test",
					},
				},
				Days:   20,
				Format: certificates.CertificateTypePrivateKey,
			},
			expectedErr: "",
		},
		{
			name: "valid-rsa-private-key",
			options: certificates.KeyPairOptions{
				Subj: certificates.Subject{
					Name: pkix.Name{
						CommonName: "test",
					},
				},
				Days:   20,
				Format: certificates.CertificateTypeRSAPrivateKey,
			},
			expectedErr: "",
		},
		{
			name:        "invalid-key-type",
			expectedErr: "Invalid certificate type CERTIFICATE",
			options: certificates.KeyPairOptions{
				Subj: certificates.Subject{
					Name: pkix.Name{
						CommonName: "test",
					},
				},
				Days:   20,
				Format: certificates.CertificateTypeCertificate,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			key, crt, err := certificates.GenerateEncodedKeyCertPair(tt.options)
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				doValidation(t, key, tt.options.Format)
				doValidation(t, crt, certificates.CertificateTypeCertificate)
			}
		})
	}
}

type testCase struct {
	name         string
	crtContent   []byte
	expectedType certificates.CertificateType
	expectErr    bool
}

var testData = map[string][]byte{
	"valid-crt": []byte(`
-----BEGIN CERTIFICATE-----
MIIFCTCCAvGgAwIBAgIUaa8FYrh81DgSJbtUMlS2IjMBIl8wDQYJKoZIhvcNAQEL
BQAwFDESMBAGA1UEAwwJaGVsbG8uY29tMB4XDTIwMTAxNDA1MDYxMVoXDTIxMTAx
NDA1MDYxMVowFDESMBAGA1UEAwwJaGVsbG8uY29tMIICIjANBgkqhkiG9w0BAQEF
AAOCAg8AMIICCgKCAgEA5lzV3zMjLRt8Cf4hKgizLanoC2ipitMrnK1eq8pW3GHx
qFCghFlUNeAxnoTa01KKYH1BdS05Gg4n4S4RgfiFAezzG6D2rLtTeyVXKmjy5Ify
GBPHSlXXbya8TgPYPGoBEwFtghiOdTOigu6wIE/Dt4AKN6D6LHPeWcKUmIR2Z62a
PW3j3p4IkW4LMLnN21eAY6P7Q9cXGUKkGkH0+7DGgmYau7cV6ZN+3b6IoC7J08o2
ziYzVJbjtqLWbHSQF5agZazJXdSnC+3DHhTbWRuavWvVcXEzHLNBeTv6jTNXpiSP
dAzrhpH0YnK7Kl8ybdXYFIOMsYClkVeboJKZ5XNW9z1WbS1kFnFP0MeuAlJvudyV
zJaOj8BgluikcdeS7KrHg6YZPS8qcW+OtO+FQv/qjskNfpM5dE6wxxR9Ph/rcPKx
WBEjHV+qPpGjXanzdCqxd/pRWWylNkb9PTkB05XX1o5ir7tYfJb26Y78GlrWte6d
cixil1chXS9PK9RyOJPZx2xBiM2s0tBM1rfksHw2YJQv2o7omfMHw253z+VoTAze
/eyChYRBNjjOfGhUVLeqOkUOtSboUOdlj5M1EiZTtO2JDyE3GewybKDlO/ICr5QF
5cOM1SzbTie2D5Z0XLJPFivbjmLwfZqExdpVITgnXYSLU6Jmh86RuTsn1EuCfssC
AwEAAaNTMFEwHQYDVR0OBBYEFPd5QlxeVPONBG/nFhVk5NJ0PYnhMB8GA1UdIwQY
MBaAFPd5QlxeVPONBG/nFhVk5NJ0PYnhMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZI
hvcNAQELBQADggIBADMKf/gDvc3ooBYudFB44wDd5d70S/NK0sOZLS9BW48ZM707
K32UeukAZ1xljP0lTF+pvww5UwhTBBahujDRzO9dstDzC+vmzcNt0X7DIkIdOIWc
sS/EkMB1vy9uc5O6c55qDM4rFDjXEqv9Vdt96kFwF2wbX/pTWKqFwFGvEoTE0QSE
z4hpCw07n8r2xyx4jEdf9LN516pd8mSu3OgG684JyOv/OTbyOJCpAzhIzPJZToJM
bwNCQ2469Ct60bNBdtEzVlPNdeL/4LJiBiscHP8a1a8YSp8Q/Tho5RZTD8N4/mwn
vAtG0MN/jWPNA0uNu6X8AXCH1CVnXHk5+uErD9j7+ZNJmsZsK54zNpBsZfUM7HQH
wrrDHM5eun8EsCXBUGlumvD0vWbXWWO3d3+cJ6WzSh00y32HVkkX5WW4Pz9olQG2
bsMXjJwwXWyxlsN30hgSOmk3cL2dWL42PEHvl33VfMMgywc5GfafgeMahvNWKzWh
sxFpYzLE2kEbPx0hWfspXS1jQrSbE7PDrWLVnTSguWuoSFDSX+Pal0HFr9vp90Hy
Fb4TpD+OodWAgwsuWo9/mgC2/CDlkh0/lTUScsk0E4pCN3zxMM2lVStNSuteQPBF
DOZaGZQMsLhmFVA0Pndgby6Jv2OkyG6NyQsO/bzhEaDG72yauK53mCpqMGc6
-----END CERTIFICATE-----
`),
	"invalid-content-crt": []byte(`
-----BEGIN CERTIFICATE-----
MIIEpQIBAAKCAQEA0JY++TITc8MYeyX7coG4iH8nbbq7tl6CzfH/yBHT7BctUoJh
dPtpwoHaHIdtlh8mIxbMy5jMjvqA94gMXAebgBT+s4tErHwJ7ETx0MrMOdCTmUjX
5z9UEIZLVO4FnP/C/Ci85kRith7xyPEgkGjxGETbezS9V2npGFox3Lo7wwHeCUqP
sffEE0E20+lb6r8SD/30S85u6NPkTMKfY3McRwBWWhwIep2U94/9aa6TXvIJQQ0j
JU0k2xb6x8zFtBQCbMXziNuFoxkCKBgBu2lq62M+bNHolZZ3CUWKuB3ToUuyrFrA
pspIOCHxhwDzU4um1QUu+2YLLGgMJa2tNOh0+QIDAQABAoIBAHzS3L6d0/w7rUPM
+AuPS5ILndnRnJHHPznlby8YVBz9xbaRpaau6ZxnvtHBzbe/zj/DXi0ctJV/nXwE
I3lTaCAe8Ekbt64M0JderuNG6S5T/nAFooaVZEY7R4t8oUlR2Sqzak/WbsgT/pdE
jTs+QcFHO50gc4qDK+XR2/L+U9MfDx6tHxNDFArBtTJDMUryXTTfpLRq6loIYhwK
P4g3HQXxPr64ek+fPf4Xm19frhPvZEna0t+MXmhiaVblrq+H36c9XEJcjMTe49lQ
p3Q61AQtnk4qe/Sueoi/u10cA3CAHaOi6rvTsxrcNSg/Lv1XxlB2U51x2KZlSzTp
UVU+FIECgYEA8Uz+k8xIewCf8rFvyJAtFmD9rMmm6Wva+fzC01iqTi22g9DX1kPw
Mi6c4LnaLhE8kz7+V+Cw7ZiB18VbTXVArdwOUH94goE1SLtLllbx2bRCAn14DWlB
tH5JehsJe4C5cwFEiV+8VYmxF/B9ZWqfKdITCvjIwlPCRj1D+RZnuMkCgYEA3UsW
9o/Qq93ldW4edJAapbcuU6LR+WV7VTCvSKQ9kQcRictmqtptgsLATk8B96hxnqch
NASgaCH4/Sr8LkMNOJgtmP/P71aiMMjaEeU2baCuWQ4pdE5XwS8H3aRL3qKP0dAe
sVKrps19ZvGCxP3W2iFrGAH6bJHF26tF/YXvIrECgYEAmppgSkYK8nRWBuNU4cYu
fTYrknepL8lhBebC1TLr+yci15YJlEj3Ls/ax8mMVxPIIfescpWOBs099AeJFjnX
9Q0XRtBFYCh1AWKvbWXLk1cBLCNDtiQIayK25TtJeg3hxCO9y97BBnUwOExnq4EC
9YKZnOAFkSylPuemE4QddLECgYEAsJVi1Y0dLof6uiING2aCXQo3ZXXfp+ta5zfa
J1Un67qAPDyayGtUR6uwWMyi/UTkpX0n+aJXfcDeNuc+JIxM2IRWnmhDPPEcq2Ea
4nzNWd2GQnoSikSZsgYdeLfJ8vY1XW99jnIxlwESuDqv5xHHiHhyRM4PTuNjx058
ozllAYECgYEA1Q9bp8kX+5rDFw12fk/5Jj+D0WlNh6rlrSjAs/xou23Ha2BefakO
G9soCxY5eYiXes+Ng7y11coUOOieNXlGHJPo9qQ1+daIB1RAKPDuzRClrLeApVq+
ut2rBZUd4Mn3zRU5aQucB70786iXIDtFw2QwBlNHglsZmYiyNhB37xQ=
-----END CERTIFICATE-----
`),
	"invalid-header": []byte(`
-----BEGIN CERT-----
MIIFCTCCAvGgAwIBAgIUaa8FYrh81DgSJbtUMlS2IjMBIl8wDQYJKoZIhvcNAQEL
BQAwFDESMBAGA1UEAwwJaGVsbG8uY29tMB4XDTIwMTAxNDA1MDYxMVoXDTIxMTAx
NDA1MDYxMVowFDESMBAGA1UEAwwJaGVsbG8uY29tMIICIjANBgkqhkiG9w0BAQEF
AAOCAg8AMIICCgKCAgEA5lzV3zMjLRt8Cf4hKgizLanoC2ipitMrnK1eq8pW3GHx
qFCghFlUNeAxnoTa01KKYH1BdS05Gg4n4S4RgfiFAezzG6D2rLtTeyVXKmjy5Ify
GBPHSlXXbya8TgPYPGoBEwFtghiOdTOigu6wIE/Dt4AKN6D6LHPeWcKUmIR2Z62a
PW3j3p4IkW4LMLnN21eAY6P7Q9cXGUKkGkH0+7DGgmYau7cV6ZN+3b6IoC7J08o2
ziYzVJbjtqLWbHSQF5agZazJXdSnC+3DHhTbWRuavWvVcXEzHLNBeTv6jTNXpiSP
dAzrhpH0YnK7Kl8ybdXYFIOMsYClkVeboJKZ5XNW9z1WbS1kFnFP0MeuAlJvudyV
zJaOj8BgluikcdeS7KrHg6YZPS8qcW+OtO+FQv/qjskNfpM5dE6wxxR9Ph/rcPKx
WBEjHV+qPpGjXanzdCqxd/pRWWylNkb9PTkB05XX1o5ir7tYfJb26Y78GlrWte6d
cixil1chXS9PK9RyOJPZx2xBiM2s0tBM1rfksHw2YJQv2o7omfMHw253z+VoTAze
/eyChYRBNjjOfGhUVLeqOkUOtSboUOdlj5M1EiZTtO2JDyE3GewybKDlO/ICr5QF
5cOM1SzbTie2D5Z0XLJPFivbjmLwfZqExdpVITgnXYSLU6Jmh86RuTsn1EuCfssC
AwEAAaNTMFEwHQYDVR0OBBYEFPd5QlxeVPONBG/nFhVk5NJ0PYnhMB8GA1UdIwQY
MBaAFPd5QlxeVPONBG/nFhVk5NJ0PYnhMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZI
hvcNAQELBQADggIBADMKf/gDvc3ooBYudFB44wDd5d70S/NK0sOZLS9BW48ZM707
K32UeukAZ1xljP0lTF+pvww5UwhTBBahujDRzO9dstDzC+vmzcNt0X7DIkIdOIWc
sS/EkMB1vy9uc5O6c55qDM4rFDjXEqv9Vdt96kFwF2wbX/pTWKqFwFGvEoTE0QSE
z4hpCw07n8r2xyx4jEdf9LN516pd8mSu3OgG684JyOv/OTbyOJCpAzhIzPJZToJM
bwNCQ2469Ct60bNBdtEzVlPNdeL/4LJiBiscHP8a1a8YSp8Q/Tho5RZTD8N4/mwn
vAtG0MN/jWPNA0uNu6X8AXCH1CVnXHk5+uErD9j7+ZNJmsZsK54zNpBsZfUM7HQH
wrrDHM5eun8EsCXBUGlumvD0vWbXWWO3d3+cJ6WzSh00y32HVkkX5WW4Pz9olQG2
bsMXjJwwXWyxlsN30hgSOmk3cL2dWL42PEHvl33VfMMgywc5GfafgeMahvNWKzWh
sxFpYzLE2kEbPx0hWfspXS1jQrSbE7PDrWLVnTSguWuoSFDSX+Pal0HFr9vp90Hy
Fb4TpD+OodWAgwsuWo9/mgC2/CDlkh0/lTUScsk0E4pCN3zxMM2lVStNSuteQPBF
DOZaGZQMsLhmFVA0Pndgby6Jv2OkyG6NyQsO/bzhEaDG72yauK53mCpqMGc6
-----END CERT-----
`),
	"valid-key-rsa": []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA0JY++TITc8MYeyX7coG4iH8nbbq7tl6CzfH/yBHT7BctUoJh
dPtpwoHaHIdtlh8mIxbMy5jMjvqA94gMXAebgBT+s4tErHwJ7ETx0MrMOdCTmUjX
5z9UEIZLVO4FnP/C/Ci85kRith7xyPEgkGjxGETbezS9V2npGFox3Lo7wwHeCUqP
sffEE0E20+lb6r8SD/30S85u6NPkTMKfY3McRwBWWhwIep2U94/9aa6TXvIJQQ0j
JU0k2xb6x8zFtBQCbMXziNuFoxkCKBgBu2lq62M+bNHolZZ3CUWKuB3ToUuyrFrA
pspIOCHxhwDzU4um1QUu+2YLLGgMJa2tNOh0+QIDAQABAoIBAHzS3L6d0/w7rUPM
+AuPS5ILndnRnJHHPznlby8YVBz9xbaRpaau6ZxnvtHBzbe/zj/DXi0ctJV/nXwE
I3lTaCAe8Ekbt64M0JderuNG6S5T/nAFooaVZEY7R4t8oUlR2Sqzak/WbsgT/pdE
jTs+QcFHO50gc4qDK+XR2/L+U9MfDx6tHxNDFArBtTJDMUryXTTfpLRq6loIYhwK
P4g3HQXxPr64ek+fPf4Xm19frhPvZEna0t+MXmhiaVblrq+H36c9XEJcjMTe49lQ
p3Q61AQtnk4qe/Sueoi/u10cA3CAHaOi6rvTsxrcNSg/Lv1XxlB2U51x2KZlSzTp
UVU+FIECgYEA8Uz+k8xIewCf8rFvyJAtFmD9rMmm6Wva+fzC01iqTi22g9DX1kPw
Mi6c4LnaLhE8kz7+V+Cw7ZiB18VbTXVArdwOUH94goE1SLtLllbx2bRCAn14DWlB
tH5JehsJe4C5cwFEiV+8VYmxF/B9ZWqfKdITCvjIwlPCRj1D+RZnuMkCgYEA3UsW
9o/Qq93ldW4edJAapbcuU6LR+WV7VTCvSKQ9kQcRictmqtptgsLATk8B96hxnqch
NASgaCH4/Sr8LkMNOJgtmP/P71aiMMjaEeU2baCuWQ4pdE5XwS8H3aRL3qKP0dAe
sVKrps19ZvGCxP3W2iFrGAH6bJHF26tF/YXvIrECgYEAmppgSkYK8nRWBuNU4cYu
fTYrknepL8lhBebC1TLr+yci15YJlEj3Ls/ax8mMVxPIIfescpWOBs099AeJFjnX
9Q0XRtBFYCh1AWKvbWXLk1cBLCNDtiQIayK25TtJeg3hxCO9y97BBnUwOExnq4EC
9YKZnOAFkSylPuemE4QddLECgYEAsJVi1Y0dLof6uiING2aCXQo3ZXXfp+ta5zfa
J1Un67qAPDyayGtUR6uwWMyi/UTkpX0n+aJXfcDeNuc+JIxM2IRWnmhDPPEcq2Ea
4nzNWd2GQnoSikSZsgYdeLfJ8vY1XW99jnIxlwESuDqv5xHHiHhyRM4PTuNjx058
ozllAYECgYEA1Q9bp8kX+5rDFw12fk/5Jj+D0WlNh6rlrSjAs/xou23Ha2BefakO
G9soCxY5eYiXes+Ng7y11coUOOieNXlGHJPo9qQ1+daIB1RAKPDuzRClrLeApVq+
ut2rBZUd4Mn3zRU5aQucB70786iXIDtFw2QwBlNHglsZmYiyNhB37xQ=
-----END RSA PRIVATE KEY-----
`),
	"valid-key": []byte(`
-----BEGIN PRIVATE KEY-----
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
reA3K5ZjwPKGF4+9Cb4/Ezx+rTETLmyU
-----END PRIVATE KEY-----
`),
	"invalid-content-key": []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIFCTCCAvGgAwIBAgIUaa8FYrh81DgSJbtUMlS2IjMBIl8wDQYJKoZIhvcNAQEL
BQAwFDESMBAGA1UEAwwJaGVsbG8uY29tMB4XDTIwMTAxNDA1MDYxMVoXDTIxMTAx
NDA1MDYxMVowFDESMBAGA1UEAwwJaGVsbG8uY29tMIICIjANBgkqhkiG9w0BAQEF
AAOCAg8AMIICCgKCAgEA5lzV3zMjLRt8Cf4hKgizLanoC2ipitMrnK1eq8pW3GHx
qFCghFlUNeAxnoTa01KKYH1BdS05Gg4n4S4RgfiFAezzG6D2rLtTeyVXKmjy5Ify
GBPHSlXXbya8TgPYPGoBEwFtghiOdTOigu6wIE/Dt4AKN6D6LHPeWcKUmIR2Z62a
PW3j3p4IkW4LMLnN21eAY6P7Q9cXGUKkGkH0+7DGgmYau7cV6ZN+3b6IoC7J08o2
ziYzVJbjtqLWbHSQF5agZazJXdSnC+3DHhTbWRuavWvVcXEzHLNBeTv6jTNXpiSP
dAzrhpH0YnK7Kl8ybdXYFIOMsYClkVeboJKZ5XNW9z1WbS1kFnFP0MeuAlJvudyV
zJaOj8BgluikcdeS7KrHg6YZPS8qcW+OtO+FQv/qjskNfpM5dE6wxxR9Ph/rcPKx
WBEjHV+qPpGjXanzdCqxd/pRWWylNkb9PTkB05XX1o5ir7tYfJb26Y78GlrWte6d
cixil1chXS9PK9RyOJPZx2xBiM2s0tBM1rfksHw2YJQv2o7omfMHw253z+VoTAze
/eyChYRBNjjOfGhUVLeqOkUOtSboUOdlj5M1EiZTtO2JDyE3GewybKDlO/ICr5QF
5cOM1SzbTie2D5Z0XLJPFivbjmLwfZqExdpVITgnXYSLU6Jmh86RuTsn1EuCfssC
AwEAAaNTMFEwHQYDVR0OBBYEFPd5QlxeVPONBG/nFhVk5NJ0PYnhMB8GA1UdIwQY
MBaAFPd5QlxeVPONBG/nFhVk5NJ0PYnhMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZI
hvcNAQELBQADggIBADMKf/gDvc3ooBYudFB44wDd5d70S/NK0sOZLS9BW48ZM707
K32UeukAZ1xljP0lTF+pvww5UwhTBBahujDRzO9dstDzC+vmzcNt0X7DIkIdOIWc
sS/EkMB1vy9uc5O6c55qDM4rFDjXEqv9Vdt96kFwF2wbX/pTWKqFwFGvEoTE0QSE
z4hpCw07n8r2xyx4jEdf9LN516pd8mSu3OgG684JyOv/OTbyOJCpAzhIzPJZToJM
bwNCQ2469Ct60bNBdtEzVlPNdeL/4LJiBiscHP8a1a8YSp8Q/Tho5RZTD8N4/mwn
vAtG0MN/jWPNA0uNu6X8AXCH1CVnXHk5+uErD9j7+ZNJmsZsK54zNpBsZfUM7HQH
wrrDHM5eun8EsCXBUGlumvD0vWbXWWO3d3+cJ6WzSh00y32HVkkX5WW4Pz9olQG2
bsMXjJwwXWyxlsN30hgSOmk3cL2dWL42PEHvl33VfMMgywc5GfafgeMahvNWKzWh
sxFpYzLE2kEbPx0hWfspXS1jQrSbE7PDrWLVnTSguWuoSFDSX+Pal0HFr9vp90Hy
Fb4TpD+OodWAgwsuWo9/mgC2/CDlkh0/lTUScsk0E4pCN3zxMM2lVStNSuteQPBF
DOZaGZQMsLhmFVA0Pndgby6Jv2OkyG6NyQsO/bzhEaDG72yauK53mCpqMGc6
-----END RSA PRIVATE KEY-----
`),
	"empty-crt": []byte(``),
}

var testCases = []testCase{
	{
		name:         "valid-crt",
		crtContent:   testData["valid-crt"],
		expectedType: certificates.CertificateTypeCertificate,
		expectErr:    false,
	},
	{
		name:         "valid-key",
		crtContent:   testData["valid-key"],
		expectedType: certificates.CertificateTypePrivateKey,
		expectErr:    false,
	},
	{
		name:         "valid-key-rsa",
		crtContent:   testData["valid-key-rsa"],
		expectedType: certificates.CertificateTypeRSAPrivateKey,
		expectErr:    false,
	},
	{
		name:         "invalid-header",
		crtContent:   testData["invalid-header"],
		expectedType: certificates.CertificateTypeUnknown,
		expectErr:    true,
	},
	{
		name:         "invalid-content-crt",
		crtContent:   testData["invalid-content-crt"],
		expectedType: certificates.CertificateTypeCertificate,
		expectErr:    true,
	},
	{
		name:         "invalid-content-key",
		crtContent:   testData["invalid-content-key"],
		expectedType: certificates.CertificateTypeRSAPrivateKey,
		expectErr:    true,
	},
	{
		name:         "empty-crt",
		crtContent:   testData["empty-crt"],
		expectedType: certificates.CertificateTypeUnknown,
		expectErr:    true,
	},
}

func TestValidateRawCertificate(t *testing.T) {
	for _, testCase := range testCases {
		receivedType, err := certificates.ValidateRawCertificate(testCase.crtContent)

		assert.Equal(t, testCase.expectedType, receivedType)
		if testCase.expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
