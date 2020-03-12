package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitDirNameFromURL(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{
			input:          "https://github.com/kubernetes/kubectl.git",
			expectedOutput: "kubectl",
		},
		{
			input:          "git@github.com:kubernetes/kubectl.git",
			expectedOutput: "kubectl",
		},
		{
			input:          "https://github.com/kubernetes/kube.somepath.ctl.git",
			expectedOutput: "kube.somepath.ctl",
		},
		{
			input:          "https://github.com/kubernetes/kubectl",
			expectedOutput: "kubectl",
		},
		{
			input:          "git@github.com:kubernetes/kubectl",
			expectedOutput: "kubectl",
		},
		{
			input:          "github.com:kubernetes/kubectl.git",
			expectedOutput: "kubectl",
		},
		{
			input:          "/kubernetes/kubectl.git",
			expectedOutput: "kubectl",
		},
		{
			input:          "/kubernetes/kubectl.git/",
			expectedOutput: "",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expectedOutput, GitDirNameFromURL(test.input))
	}
}
