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
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"golang.org/x/crypto/ssh"
)

func toUint32(i int) uint32 { return uint32(i) }

type dnParser struct {
	in    string
	i     int
	cur   bytes.Buffer
	state int
	dn    []string
}

type sshKey struct {
	Private string
	Public  string
}

func (p *dnParser) startOver() {
	p.dn = append(p.dn, p.cur.String())
	p.cur = bytes.Buffer{}
	p.state = 0
}

func (p *dnParser) parseParam(r rune) error {
	switch r {
	case '\\':
		p.state = 2
	case '+', '/':
		return fmt.Errorf("string %s has separator '%c', but didn't have value on position %d", p.in, r, p.i+1)
	case '=':
		p.state = 1
		p.cur.WriteRune(r)
	case '"', ',', '<', '>', ';':
		return fmt.Errorf("string %s position %d: having %c without '\\'", p.in, p.i+1, r)
	default:
		p.cur.WriteRune(r)
	}
	return nil
}

func (p *dnParser) parseValue(r rune) error {
	switch r {
	case '\\':
		p.state = 3
	case '+', '/':
		p.startOver()
	case '=':
		return fmt.Errorf("string %s has extra '=' on position %d", p.in, p.i+1)
	case '"', ',', '<', '>', ';':
		return fmt.Errorf("string %s position %d: having %c without '\\'", p.in, p.i+1, r)
	default:
		p.cur.WriteRune(r)
	}
	return nil
}

func (p *dnParser) parseParamEscape(r rune) error {
	switch r {
	case '=', '+', '/', '"', ',', '<', '>', ';':
		p.cur.WriteRune(r)
		p.state = 0
	default:
		return fmt.Errorf("string %s pos %d: %c shouldn't follow after '\\'", p.in, p.i+1, r)
	}
	return nil
}

func (p *dnParser) parseValueEscape(r rune) error {
	switch r {
	case '=', '+', '/', '"', ',', '<', '>', ';':
		p.cur.WriteRune(r)
		p.state = 1
	default:
		return fmt.Errorf("string %s pos %d: %c shouldn't follow after '\\'", p.in, p.i+1, r)
	}
	return nil
}

func (p *dnParser) Parse(in string) error {
	p.cur = bytes.Buffer{}
	p.state = 0
	p.dn = nil

	p.in = in
	var err error
	for p.i = 0; p.i < len(p.in); {
		r, size := utf8.DecodeRuneInString(p.in[p.i:])
		switch p.state {
		case 0: // initial state
			err = p.parseParam(r)
		case 1: // the same, but after =
			err = p.parseValue(r)
		case 2: // state inside \
			err = p.parseParamEscape(r)
		case 3: // state inside \ after =
			err = p.parseValueEscape(r)
		}

		if err != nil {
			return err
		}

		p.i += size
	}

	if p.state != 1 {
		return fmt.Errorf("string %s terminates incorrectly", p.in)
	}
	p.startOver()
	return nil
}

// Converts RFC 2253 Distinguished Names syntax back to
// Name similar to what openssl parse_name function [1]
// does, except that if it doesn't have / as a first simbol
// it assumes that the whole string is CN.
// we don't support MultiRdn - + is treated the same way as /.
// [1] https://github.com/openssl/openssl/blob/d8ab30be9cc4d4e77008d4037e696bc41ce293f8/apps/lib/apps.c#L1624
func nameFromString(in string) (*pkix.Name, error) {
	if len(in) > 0 && in[0] != '/' {
		return &pkix.Name{
			CommonName: in,
		}, nil
	}

	in = in[1:]
	if len(in) == 0 {
		return &pkix.Name{
			CommonName: in,
		}, nil
	}

	p := &dnParser{}
	err := p.Parse(in)
	if err != nil {
		return nil, err
	}

	return nameFromDn(p.dn)
}

func nameFromDn(dn []string) (*pkix.Name, error) {
	name := pkix.Name{}

	for _, v := range dn {
		sv := strings.Split(v, "=")
		if len(sv) != 2 {
			return nil, fmt.Errorf("%s must have a form key=value", v)
		}
		switch sv[0] {
		case "CN":
			if name.CommonName != "" {
				return nil, fmt.Errorf("CN is already set")
			}
			name.CommonName = sv[1]
		case "SERIALNUMBER":
			if name.SerialNumber != "" {
				return nil, fmt.Errorf("SERIALNUMBER is already set")
			}
			name.SerialNumber = sv[1]
		case "C":
			name.Country = append(name.Country, sv[1])
		case "O":
			name.Organization = append(name.Organization, sv[1])
		case "OU":
			name.OrganizationalUnit = append(name.OrganizationalUnit, sv[1])
		case "L":
			name.Locality = append(name.Locality, sv[1])
		case "ST":
			name.Province = append(name.Province, sv[1])
		case "STREET":
			name.StreetAddress = append(name.StreetAddress, sv[1])
		case "POSTALCODE":
			name.PostalCode = append(name.PostalCode, sv[1])
		default:
			return nil, fmt.Errorf("unsupported property %s", sv[0])
		}
	}

	return &name, nil
}

func generateCertificateAuthorityEx(
	subj string,
	daysValid int,
) (certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return certificate{}, fmt.Errorf("error generating rsa key: %s", err)
	}

	return generateCertificateAuthorityWithKeyInternalEx(subj, daysValid, priv)
}

// genSSHKeyPair make a pair of public and private keys for SSH access.
// Public key is encoded in the format for inclusion in an OpenSSH authorized_keys file.
// Private Key generated is PEM encoded
func genSSHKeyPair(encryptionBit int) (sshKey, error) {
	key := sshKey{}
	privateKey, err := rsa.GenerateKey(rand.Reader, encryptionBit)
	if err != nil {
		return key, err
	}

	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	var private bytes.Buffer
	if err = pem.Encode(&private, privateKeyPEM); err != nil {
		return key, err
	}

	// generate public key
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return key, err
	}

	public := ssh.MarshalAuthorizedKey(pub)
	key.Public = string(public)
	key.Private = private.String()
	return key, nil
}

func generateCertificateAuthorityWithPEMKeyEx(
	subj string,
	daysValid int,
	privPEM string,
) (certificate, error) {
	priv, err := parsePrivateKeyPEM(privPEM)
	if err != nil {
		return certificate{}, fmt.Errorf("parsing private key: %s", err)
	}
	return generateCertificateAuthorityWithKeyInternalEx(subj, daysValid, priv)
}

func generateCertificateAuthorityWithKeyInternalEx(
	subj string,
	daysValid int,
	priv crypto.PrivateKey,
) (certificate, error) {
	ca := certificate{}

	template, err := getBaseCertTemplateEx(subj, nil, nil, daysValid)
	if err != nil {
		return ca, err
	}
	// Override KeyUsage and IsCA
	template.KeyUsage = x509.KeyUsageKeyEncipherment |
		x509.KeyUsageDigitalSignature |
		x509.KeyUsageCertSign
	template.IsCA = true

	ca.Cert, ca.Key, err = getCertAndKey(template, priv, template, priv)

	return ca, err
}

func generateSignedCertificateEx(
	subj string,
	ips []interface{},
	alternateDNS []interface{},
	daysValid int,
	ca certificate,
) (certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return certificate{}, fmt.Errorf("error generating rsa key: %s", err)
	}
	return generateSignedCertificateWithKeyInternalEx(subj, ips, alternateDNS, daysValid, ca, priv)
}

func generateSignedCertificateWithPEMKeyEx(
	subj string,
	ips []interface{},
	alternateDNS []interface{},
	daysValid int,
	ca certificate,
	privPEM string,
) (certificate, error) {
	priv, err := parsePrivateKeyPEM(privPEM)
	if err != nil {
		return certificate{}, fmt.Errorf("parsing private key: %s", err)
	}
	return generateSignedCertificateWithKeyInternalEx(subj, ips, alternateDNS, daysValid, ca, priv)
}

func generateSignedCertificateWithKeyInternalEx(
	subj string,
	ips []interface{},
	alternateDNS []interface{},
	daysValid int,
	ca certificate,
	priv crypto.PrivateKey,
) (certificate, error) {
	cert := certificate{}

	decodedSignerCert, _ := pem.Decode([]byte(ca.Cert))
	if decodedSignerCert == nil {
		return cert, errors.New("unable to decode certificate")
	}
	signerCert, err := x509.ParseCertificate(decodedSignerCert.Bytes)
	if err != nil {
		return cert, fmt.Errorf(
			"error parsing certificate: decodedSignerCert.Bytes: %s",
			err,
		)
	}
	signerKey, err := parsePrivateKeyPEM(ca.Key)
	if err != nil {
		return cert, fmt.Errorf(
			"error parsing private key: %s",
			err,
		)
	}

	template, err := getBaseCertTemplateEx(subj, ips, alternateDNS, daysValid)
	if err != nil {
		return cert, err
	}

	cert.Cert, cert.Key, err = getCertAndKey(
		template,
		priv,
		signerCert,
		signerKey,
	)

	return cert, err
}

func getBaseCertTemplateEx(
	subj string,
	ips []interface{},
	alternateDNS []interface{},
	daysValid int,
) (*x509.Certificate, error) {
	ipAddresses, err := getNetIPs(ips)
	if err != nil {
		return nil, err
	}
	dnsNames, err := getAlternateDNSStrs(alternateDNS)
	if err != nil {
		return nil, err
	}
	serialNumberUpperBound := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberUpperBound)
	if err != nil {
		return nil, err
	}
	name, err := nameFromString(subj)
	if err != nil {
		return nil, err
	}
	return &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      *name,
		IPAddresses:  ipAddresses,
		DNSNames:     dnsNames,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24 * time.Duration(daysValid)),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
	}, nil
}
