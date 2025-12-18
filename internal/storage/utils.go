package storage

import (
	"os"
)

// ensureDir creates a directory if it doesn't exist
func ensureDir(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}
