package util

import (
	"io"
	"os"
	"path/filepath"
)

// copyFile handles the file copying operation, creating parent directories if needed
func CopyFile(src, dst string) error {
	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Open source file
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// Create destination file
	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	// Copy the contents
	_, err = io.Copy(destination, source)
	return err
}
