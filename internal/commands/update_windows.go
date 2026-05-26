package commands

import (
	"context"
	"os"
	"os/exec"
)

func runUpdate(ctx context.Context, c Context) (int, error) {
	// On Windows, try winget first, then fall back to reinstall
	if winget, err := exec.LookPath("winget"); err == nil {
		cmd := exec.CommandContext(ctx, winget, "upgrade", "jolehuit.clother")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			if exit, ok := err.(*exec.ExitError); ok {
				return exit.ExitCode(), nil
			}
			return 1, err
		}
		return 0, nil
	}

	// Fallback: just run install which will download the latest
	return runInstall(ctx, c)
}
