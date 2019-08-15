package secret

import (
	crypto "crypto/rand"
	"encoding/binary"
	"math/rand"

	"opendev.org/airship/airshipctl/pkg/log"
)

// Source implements rand.Source
type Source struct{}

var _ rand.Source = &Source{}

// Uint64 returns a secure random uint64 in the range [0, 1<<64]. It will fail
// if an error is returned from the system's secure random numer generator
func (s *Source) Uint64() uint64 {
	var value uint64
	err := binary.Read(crypto.Reader, binary.BigEndian, &value)
	if err != nil {
		log.Fatalf("could not generate a random number: %s", err.Error())
	}
	return value
}

// Int63 returns a secure random int64 in the range [0, 1<<63]
func (s *Source) Int63() int64 {
	return int64(s.Uint64() & ^(uint64(1 << 63)))
}

// Seed does nothing, since Source will use the crypto library
func (s *Source) Seed(_ int64) {}
