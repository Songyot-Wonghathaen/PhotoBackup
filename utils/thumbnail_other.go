//go:build !windows

package utils

import (
	"os"
)

// GetThumbnailBytes fallback for non-Windows platforms
func GetThumbnailBytes(filePath string, maxDim int) ([]byte, error) {
	return os.ReadFile(filePath)
}
