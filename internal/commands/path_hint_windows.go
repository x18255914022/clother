package commands

import "fmt"

func pathHint(binDir string) string {
	return fmt.Sprintf("add `%s` to your PATH environment variable (System Properties > Environment Variables)", binDir)
}
