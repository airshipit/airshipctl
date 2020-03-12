package util

import (
	"path/filepath"
	"strings"
)

// GitDirNameFromURL extract directory name of the repository from URL
func GitDirNameFromURL(urlString string) string {
	_, fileName := filepath.Split(urlString)
	return strings.TrimSuffix(fileName, ".git")
}
