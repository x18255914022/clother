//go:build !windows

package config

import "os"

func setFilePermissions(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}
