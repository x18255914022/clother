package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/jolehuit/clother/internal/cli"
	"github.com/jolehuit/clother/internal/profiles"
)

func Dispatch(ctx context.Context, c Context, command string, args []string) (int, error) {
	switch command {
	case "":
		cli.ShowBrief(c.Output.Stdout)
		return 0, nil
	case "config":
		return runConfig(ctx, c, args)
	case "list":
		return runList(ctx, c)
	case "info":
		return runInfo(ctx, c, args)
	case "test":
		return runTest(ctx, c, args)
	case "status":
		return runStatus(ctx, c)
	case "install":
		return runInstall(ctx, c)
	case "update":
		return runUpdate(ctx, c)
	case "uninstall":
		return runUninstall(ctx, c)
	case "help":
		cli.ShowFull(c.Output.Stdout, c.Catalog)
		return 0, nil
	default:
		// Check if command is a provider name
		if target, err := resolveProviderCommand(c, command, args); err == nil {
			launcherArgs, noBanner := cli.ParseLauncherArgs(args)
			return RunLauncher(ctx, c.Paths, c.Secrets, target, launcherArgs, noBanner)
		}
		return 1, fmt.Errorf("unknown command %q", command)
	}
}

// resolveProviderCommand checks if the command is a provider name and resolves it.
// Handles: clother zai, clother or kimi-k25, clother custom myprovider
func resolveProviderCommand(c Context, command string, args []string) (profiles.Target, error) {
	profile := command
	remaining := args

	// Handle gateway commands: clother or <alias>, clother custom <name>
	if profile == "or" {
		if len(remaining) == 0 || strings.HasPrefix(remaining[0], "-") {
			return profiles.Target{}, fmt.Errorf("usage: clother or <alias> [args...]")
		}
		profile = "or-" + remaining[0]
		remaining = remaining[1:]
	} else if profile == "custom" {
		if len(remaining) == 0 || strings.HasPrefix(remaining[0], "-") {
			return profiles.Target{}, fmt.Errorf("usage: clother custom <provider-name> [args...]")
		}
		name := remaining[0]
		if _, ok := c.Config.CustomProviders[name]; !ok {
			return profiles.Target{}, fmt.Errorf("unknown custom provider %q", name)
		}
		profile = name
		remaining = remaining[1:]
	}

	target, err := profiles.Resolve(profile, c.Catalog, c.Config)
	if err != nil {
		return profiles.Target{}, err
	}
	return target, nil
}
