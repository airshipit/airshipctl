package document

import (
	b64 "encoding/base64"
)

// GetSecretDataKey understands how to retrieve a specific top level key from a secret
// that may have the data stored under a data or stringData field in which
// case the key may be base64 encoded or it may be plain text
//
// it is meant to be used by other high level dochelpers
func GetSecretDataKey(cfg Document, key string) (string, error) {
	var needsBase64Decode = true
	var docName = cfg.GetName()

	// this purposely doesn't handle binaryData as that isn't
	// something we could support anyways
	data, err := cfg.GetStringMap("stringData")
	if err == nil {
		needsBase64Decode = false
	} else {
		data, err = cfg.GetStringMap("data")
		if err != nil {
			return "", ErrDocumentMalformed{
				DocName: docName,
				Message: "The secret document lacks a data or stringData top level field",
			}
		}
	}

	res, ok := data[key]
	if !ok {
		return "", ErrDocumentDataKeyNotFound{DocName: docName, Key: key}
	}

	if needsBase64Decode {
		byteSlice, err := b64.StdEncoding.DecodeString(res)
		if err != nil {
			return "", err
		}
		return string(byteSlice), nil
	}
	return res, nil
}
