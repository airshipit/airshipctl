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
	crypto "crypto/rand"
	"encoding/binary"
	"math/rand"

	"opendev.org/airship/airshipctl/pkg/log"
)

// Source implements rand.Source
type Source struct{}

var _ rand.Source = &Source{}

// Uint64 returns a secure random uint64 in the range [0, 1<<64]. It will fail
// if an error is returned from the system's secure random number generator
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
