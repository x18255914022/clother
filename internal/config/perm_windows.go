package config

import "os"

// On Windows, file permissions work differently. We skip chmod.
func setFilePermissions(path string, mode os.FileMode) error {
	return nil
}
