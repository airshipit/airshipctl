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

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"opendev.org/airship/airshipctl/pkg/log"
)

// CertificateType -list of currently supported types
type CertificateType string

const (
	// CertificateTypeCertificate  for crt files
	CertificateTypeCertificate CertificateType = "CERTIFICATE"
	// CertificateTypeRSAPrivateKey for generating PKCS1 keys
	CertificateTypeRSAPrivateKey CertificateType = "RSA PRIVATE KEY"
	// CertificateTypePrivateKey for generating PKCS8 keys
	CertificateTypePrivateKey CertificateType = "PRIVATE KEY"
	// CertificateTypeUnknown for unknown types
	CertificateTypeUnknown CertificateType = "UNKNOWN"
)

// KeyPairOptions holds options for key-pair generation
type KeyPairOptions struct {
	Subj   Subject
	Format CertificateType
	Days   int
}

// Subject data for the certificate
type Subject struct {
	pkix.Name
}

// GenerateEncodedKeyCertPair generates PEM encoded RSA key and crt data
func GenerateEncodedKeyCertPair(options KeyPairOptions) (caPrivKeyPEM string, caPEM string, err error) {
	if options.Format != CertificateTypePrivateKey && options.Format != CertificateTypeRSAPrivateKey {
		log.Printf("Received unknown Private Key type %s", options.Format)
		return "", "", ErrInvalidType{inputType: string(options.Format)}
	}

	caPrivKey, caBytes, err := generatePrivateKeyAndCertificate(options)
	if err != nil {
		return "", "", err
	}
	caPEM, err = encodeToPem(CertificateTypeCertificate, caBytes)
	if err != nil {
		return "", "", err
	}

	var pemData []byte
	switch options.Format {
	case CertificateTypePrivateKey:
		// Mainly used for creating k8s CA, Etcd and Proxy certificates
		pemData, err = x509.MarshalPKCS8PrivateKey(caPrivKey)
		if err != nil {
			return "", "", err
		}
	case CertificateTypeRSAPrivateKey:
		// Mainly used for creating k8s Service Account certificates
		pemData = x509.MarshalPKCS1PrivateKey(caPrivKey)
	}

	caPrivKeyPEM, err = encodeToPem(options.Format, pemData)
	if err != nil {
		return "", "", err
	}

	return caPrivKeyPEM, caPEM, nil
}

// generatePrivateKeyAndCertificate generates Private Key and Certificate
func generatePrivateKeyAndCertificate(options KeyPairOptions) (caPrivKey *rsa.PrivateKey,
	caBytes []byte, err error) {
	caPrivKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	ca := &x509.Certificate{
		SerialNumber:          big.NewInt(2020),
		Subject:               options.Subj.Name,
		NotAfter:              time.Now().AddDate(0, 0, options.Days),
		NotBefore:             time.Now(),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	caBytes, err = x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	return caPrivKey, caBytes, nil
}

// encodeToPem encodes the data to PEM
func encodeToPem(encodeType CertificateType, data []byte) (string, error) {
	encodedData := bytes.NewBuffer([]byte{})

	err := pem.Encode(encodedData, &pem.Block{
		Type:    string(encodeType),
		Bytes:   data,
		Headers: make(map[string]string),
	})

	return encodedData.String(), err
}

// ValidateRawCertificate checks if the crt/key is valid
func ValidateRawCertificate(certByte []byte) (CertificateType, error) {
	pemBlock, _ := pem.Decode(certByte)

	if pemBlock == nil {
		return CertificateTypeUnknown, ErrMalformedCertificateData{errMsg: "decoding of the certificate failed"}
	}
	switch pemBlock.Type {
	case string(CertificateTypeRSAPrivateKey):
		_, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
		return CertificateTypeRSAPrivateKey, err
	case string(CertificateTypePrivateKey):
		_, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
		return CertificateTypePrivateKey, err
	case string(CertificateTypeCertificate):
		_, err := x509.ParseCertificate(pemBlock.Bytes)
		return CertificateTypeCertificate, err
	default:
		return CertificateTypeUnknown, ErrInvalidType{inputType: pemBlock.Type}
	}
}
