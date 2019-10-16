package log_test

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/log"
)

var logFormatRegex = regexp.MustCompile(`^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} .*`)

const prefixLength = len("2001/02/03 16:05:06 ")

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
		actual = actual[prefixLength:]
		assert.Equal(expected, actual)
	})

	t.Run("DebugfTrue", func(t *testing.T) {
		output := new(bytes.Buffer)
		log.Init(true, output)

		log.Debugf("%s %d", "DebugfTrue args", 5)
		actual := output.String()

		expected := "DebugfTrue args 5\n"
		require.Regexp(logFormatRegex, actual)
		actual = actual[prefixLength:]
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
