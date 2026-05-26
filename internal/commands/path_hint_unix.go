//go:build !windows

package commands

import "fmt"

func pathHint(binDir string) string {
	return fmt.Sprintf("add `export PATH=\"%s:$PATH\"` to your shell profile and restart your shell", binDir)
}
