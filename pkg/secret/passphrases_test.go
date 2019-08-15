package secret_test

import (
	"math/rand"
	"strings"
	"testing"

	"opendev.org/airship/airshipctl/pkg/secret"
)

const (
	asciiLowers  = "abcdefghijklmnopqrstuvwxyz"
	asciiUppers  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	asciiNumbers = "0123456789"
	asciiSymbols = "@#&-+=?"

	defaultLength = 24
)

func TestDeterministicGenerateValidPassphrase(t *testing.T) {
	testSource := rand.NewSource(42)
	engine := secret.NewPassphraseEngine(testSource)

	// pre-calculated for rand.NewSource(42)
	expectedPassphrases := []string{
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

	for i, expected := range expectedPassphrases {
		actual := engine.GeneratePassphrase()
		if expected != actual {
			t.Errorf("Call #%d to engine.GeneratePassphrase() should have returned %s, got %s",
				i, expected, actual)
		}
	}
}

func TestNondeterministicGenerateValidPassphrase(t *testing.T) {
	// Due to the nondeterminism of random number generators, this
	// functionality is impossible to fully test. Let's just generate
	// enough passphrases that we can be confident in the randomness.
	// Fortunately, Go is pretty fast, so we can do upward of 100,000 of
	// these without slowing down the test too much
	engine := secret.NewPassphraseEngine(nil)
	for i := 0; i < 100000; i++ {
		passphrase := engine.GeneratePassphrase()
		if !isValid(passphrase) {
			t.Errorf("The engine generated an invalid password: %s", passphrase)
		}
	}
}

func TestGenerateValidPassphraseN(t *testing.T) {
	testSource := rand.NewSource(42)
	engine := secret.NewPassphraseEngine(testSource)
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
		passphrase := engine.GeneratePassphraseN(tt.inputLegth)
		if len(passphrase) != tt.expectedLength {
			t.Errorf(`Passphrase "%s" should have length %d, got %d\n`,
				passphrase, len(passphrase), tt.expectedLength)
		}
	}
}

func isValid(passphrase string) bool {
	if len(passphrase) < defaultLength {
		return false
	}

	charSets := []string{
		asciiLowers,
		asciiUppers,
		asciiNumbers,
		asciiSymbols,
	}
	for _, charSet := range charSets {
		if !strings.ContainsAny(passphrase, charSet) {
			return false
		}
	}
	return true
}
