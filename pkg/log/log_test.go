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

package log_test

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/log"
)

var logFormatRegex = regexp.MustCompile(`^\[airshipctl\] \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} .*`)

const prefixLength = len("[airshipctl] 2001/02/03 16:05:06 ")

func TestLoggingPrintf(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	t.Run("Print", func(t *testing.T) {
		output := new(bytes.Buffer)
		log.Init(false, output)

		log.Print("Print args ", 5)
		actual := output.String()

		expected := "Print args 5\n"
		require.Regexp(logFormatRegex, actual)
		actual = actual[prefixLength:]
		assert.Equal(expected, actual)
	})

	t.Run("Printf", func(t *testing.T) {
		output := new(bytes.Buffer)
		log.Init(false, output)

		log.Printf("%s %d", "Printf args", 5)
		actual := output.String()

		expected := "Printf args 5\n"
		require.Regexp(logFormatRegex, actual)
		actual = actual[prefixLength:]
		assert.Equal(expected, actual)
	})
}

func TestLoggingDebug(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	t.Run("DebugTrue", func(t *testing.T) {
		output := new(bytes.Buffer)
		log.Init(true, output)

		log.Debug("DebugTrue args ", 5)
		actual := output.String()

		expected := "DebugTrue args 5\n"
		require.Regexp(logFormatRegex, actual)
		lastIndex := strings.LastIndex(actual, ":")
		actual = actual[lastIndex+2:]
		assert.Equal(expected, actual)
	})

	t.Run("DebugfTrue", func(t *testing.T) {
		output := new(bytes.Buffer)
		log.Init(true, output)

		log.Debugf("%s %d", "DebugfTrue args", 5)
		actual := output.String()

		expected := "DebugfTrue args 5\n"
		require.Regexp(logFormatRegex, actual)
		lastIndex := strings.LastIndex(actual, ":")
		actual = actual[lastIndex+2:]
		assert.Equal(expected, actual)
	})

	t.Run("DebugFalse", func(t *testing.T) {
		output := new(bytes.Buffer)
		log.Init(false, output)

		log.Debug("DebugFalse args ", 5)
		assert.Equal("", output.String())
	})

	t.Run("DebugfFalse", func(t *testing.T) {
		output := new(bytes.Buffer)
		log.Init(false, output)

		log.Debugf("%s %d", "DebugFalse args", 5)
		assert.Equal("", output.String())
	})
}
