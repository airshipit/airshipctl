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

package generate

import (
	"math/rand"
	"strings"
)

const (
	defaultLength = 24

	asciiLowers  = "abcdefghijklmnopqrstuvwxyz"
	asciiUppers  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	asciiNumbers = "0123456789"
	asciiSymbols = "@#&-+=?"
)

var (
	// pool is the complete collection of characters that can be used for a
	// passphrase
	pool []byte
)

func init() {
	pool = append(pool, []byte(asciiLowers)...)
	pool = append(pool, []byte(asciiUppers)...)
	pool = append(pool, []byte(asciiNumbers)...)
	pool = append(pool, []byte(asciiSymbols)...)
}

// EncryptionKeyEngine is used to generate secure random passphrases
type EncryptionKeyEngine struct {
	rng  *rand.Rand
	pool []byte
}

// NewEncryptionKeyEngine creates an PassphraseEngine using src. If src is nil,
// the returned PassphraseEngine will use the default Source
func NewEncryptionKeyEngine(src rand.Source) *EncryptionKeyEngine {
	if src == nil {
		src = &Source{}
	}
	return &EncryptionKeyEngine{
		rng:  rand.New(src),
		pool: pool,
	}
}

// GenerateEncryptionKey returns a secure random string of length 24,
// containing at least one of each from the following sets:
// [0-9]
// [a-z]
// [A-Z]
// [@#&-+=?]
func (e *EncryptionKeyEngine) GenerateEncryptionKey() string {
	return e.GenerateEncryptionKeyN(defaultLength)
}

// GenerateEncryptionKeyN returns a secure random string containing at least
// one of each from the following sets. Its length will be max(length, 24)
// [0-9]
// [a-z]
// [A-Z]
// [@#&-+=?]
func (e *EncryptionKeyEngine) GenerateEncryptionKeyN(length int) string {
	if length < defaultLength {
		length = defaultLength
	}
	var encryptionkey string
	for !e.isValidEncryptionKey(encryptionkey) {
		var sb strings.Builder
		for i := 0; i < length; i++ {
			randIndex := e.rng.Intn(len(e.pool))
			randChar := e.pool[randIndex]
			sb.WriteString(string(randChar))
		}
		encryptionkey = sb.String()
	}
	return encryptionkey
}

func (e *EncryptionKeyEngine) isValidEncryptionKey(encryptionkey string) bool {
	if len(encryptionkey) < defaultLength {
		return false
	}

	charSets := []string{
		asciiLowers,
		asciiUppers,
		asciiNumbers,
		asciiSymbols,
	}
	for _, charSet := range charSets {
		if !strings.ContainsAny(encryptionkey, charSet) {
			return false
		}
	}
	return true
}
