package commands

import (
	"context"
	"io"
	"testing"

	"github.com/jolehuit/clother/internal/config"
	"github.com/jolehuit/clother/internal/providers"
	"github.com/jolehuit/clother/internal/ui"
)

func TestDispatchProviderSubcommand(t *testing.T) {
	catalog, err := providers.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.File{
		Version:           1,
		ProviderOverrides: map[string]config.ProviderOverride{},
		OpenRouterAliases: map[string]string{"kimi": "moonshotai/kimi-k2.5"},
		CustomProviders: map[string]config.CustomProvider{
			"myprovider": {
				Name:        "myprovider",
				DisplayName: "My Provider",
				BaseURL:     "https://example.com/anthropic",
				APIKeyEnv:   "MYPROVIDER_API_KEY",
			},
		},
	}

	tests := []struct {
		name    string
		command string
		args    []string
		wantErr bool
	}{
		{
			name:    "builtin provider",
			command: "zai",
			args:    []string{"--model", "glm-5"},
			wantErr: false,
		},
		{
			name:    "openrouter alias",
			command: "or",
			args:    []string{"kimi", "--model", "glm-5"},
			wantErr: false,
		},
		{
			name:    "custom provider",
			command: "custom",
			args:    []string{"myprovider"},
			wantErr: false,
		},
		{
			name:    "unknown provider",
			command: "unknown-provider",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "or without alias",
			command: "or",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "custom without name",
			command: "custom",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := Context{
				Config:  cfg,
				Catalog: catalog,
				Output:  &ui.Output{Stdout: io.Discard, Stderr: io.Discard},
			}

			// We can't actually run the launcher (it would exec claude),
			// so we just test that resolveProviderCommand works correctly.
			_, err := resolveProviderCommand(ctx, tt.command, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveProviderCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDispatchUnknownCommand(t *testing.T) {
	catalog, err := providers.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.File{
		Version:           1,
		ProviderOverrides: map[string]config.ProviderOverride{},
		OpenRouterAliases: map[string]string{},
		CustomProviders:   map[string]config.CustomProvider{},
	}

	ctx := context.Background()
	c := Context{
		Config:  cfg,
		Catalog: catalog,
		Output:  &ui.Output{Stdout: io.Discard, Stderr: io.Discard},
	}

	code, err := Dispatch(ctx, c, "totally-unknown", []string{})
	if err == nil {
		t.Error("expected error for unknown command")
	}
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}
