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

package generate_test

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"opendev.org/airship/airshipctl/pkg/secret/generate"
)

const (
	asciiLowers  = "abcdefghijklmnopqrstuvwxyz"
	asciiUppers  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	asciiNumbers = "0123456789"
	asciiSymbols = "@#&-+=?"

	defaultLength = 24
)

func TestDeterministicGenerateValidEncryptionKey(t *testing.T) {
	assert := assert.New(t)
	testSource := rand.NewSource(42)
	engine := generate.NewEncryptionKeyEngine(testSource)

	// pre-calculated for rand.NewSource(42)
	expectedEncryptionKey := []string{
		"erx&97vfqd7LN3HJ?t@oPhds",
		"##Xeuvf5Njy@hNWSaRoleFkf",
		"jB=kirg7acIt-=fx1Fb-tZ+7",
		"eOS#W8yoAljSThPL2oT&aUZu",
		"vlaQqKr-jXSCJfXYnvGik3b1",
		"rBKtHZkOmFUM75?c2UWiZjdh",
		"9g?QV?w6BCWN2EKAc+dZ-Jun",
		"X@IIyqAg7Mz@Wm8eRE6KMEf3",
		"7JpQkLd3ufhj4bLB8S=ipjNP",
		"XC?bDaHTa3mrBYLMG@#B=Q0B",
	}

	for i, expected := range expectedEncryptionKey {
		actual := engine.GenerateEncryptionKey()
		assert.Equal(expected, actual, "Test #%d failed", i)
	}
}

func TestNondeterministicGenerateValidEncryptionKey(t *testing.T) {
	assert := assert.New(t)
	// Due to the nondeterminism of random number generators, this
	// functionality is impossible to fully test. Let's just generate
	// enough passphrases that we can be confident in the randomness.
	// Fortunately, Go is pretty fast, so we can do upward of 10,000 of
	// these without slowing down the test too much
	charSets := []string{
		asciiLowers,
		asciiUppers,
		asciiNumbers,
		asciiSymbols,
	}

	engine := generate.NewEncryptionKeyEngine(nil)
	for i := 0; i < 10000; i++ {
		passphrase := engine.GenerateEncryptionKey()
		assert.Truef(len(passphrase) >= defaultLength,
			"%s does not meet the length requirement", passphrase)

		for _, charSet := range charSets {
			assert.Truef(strings.ContainsAny(passphrase, charSet),
				"%s does not contain any characters from [%s]",
				passphrase, charSet)
		}
	}
}

func TestGenerateValidEncryptionKeyN(t *testing.T) {
	assert := assert.New(t)
	testSource := rand.NewSource(42)
	engine := generate.NewEncryptionKeyEngine(testSource)
	tests := []struct {
		inputLegth     int
		expectedLength int
	}{
		{
			inputLegth:     10,
			expectedLength: defaultLength,
		},
		{
			inputLegth:     -5,
			expectedLength: defaultLength,
		},
		{
			inputLegth:     30,
			expectedLength: 30,
		},
	}

	for _, tt := range tests {
		passphrase := engine.GenerateEncryptionKeyN(tt.inputLegth)
		assert.Len(passphrase, tt.expectedLength)
	}
}
