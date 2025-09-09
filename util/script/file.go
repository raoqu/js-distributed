package script

import (
	"fmt"
	"os"
)

// ReadFile reads a file and returns its content as a string
func ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return string(content), nil
}
