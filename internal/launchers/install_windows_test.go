package launchers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jolehuit/clother/internal/config"
	"github.com/jolehuit/clother/internal/providers"
)

func TestSyncCreatesBinaryAndLaunchers(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	execPath := filepath.Join(root, "clother-bin.exe")
	if err := os.WriteFile(execPath, []byte("fake binary"), 0o755); err != nil {
		t.Fatal(err)
	}

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
				DisplayName: "myprovider",
				BaseURL:     "https://example.com/anthropic",
				APIKeyEnv:   "MYPROVIDER_API_KEY",
			},
		},
	}
	paths := config.Paths{
		ConfigDir:       filepath.Join(root, "config"),
		DataDir:         filepath.Join(root, "data"),
		CacheDir:        filepath.Join(root, "cache"),
		BinDir:          filepath.Join(root, "bin"),
		ManifestFile:    filepath.Join(root, "data", "launchers.json"),
		SessionPatchDir: filepath.Join(root, "data", "session-patches"),
	}

	if err := Sync(execPath, paths, catalog, cfg, false); err != nil {
		t.Fatal(err)
	}

	// Check that .bat files are created for providers
	for _, name := range []string{"clother-zai", "clother-native", "clother-or-kimi", "clother-myprovider", "clother-or", "clother-custom", "claude"} {
		batPath := filepath.Join(paths.BinDir, name+".bat")
		if _, err := os.Stat(batPath); err != nil {
			t.Fatalf("missing %s.bat: %v", name, err)
		}
	}

	// Check that clother.exe is copied
	if _, err := os.Stat(filepath.Join(paths.BinDir, "clother.exe")); err != nil {
		t.Fatal("clother.exe should be copied in normal mode")
	}
}

func TestSyncHomebrewSkipsCopyAndUsesAbsolutePaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// Simulate the Homebrew-managed binary (not in BinDir)
	homebrewBin := filepath.Join(root, "homebrew", "bin", "clother.exe")
	if err := os.MkdirAll(filepath.Dir(homebrewBin), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(homebrewBin, []byte("fake binary"), 0o755); err != nil {
		t.Fatal(err)
	}

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
	paths := config.Paths{
		ConfigDir:       filepath.Join(root, "config"),
		DataDir:         filepath.Join(root, "data"),
		CacheDir:        filepath.Join(root, "cache"),
		BinDir:          filepath.Join(root, "bin"),
		ManifestFile:    filepath.Join(root, "data", "launchers.json"),
		SessionPatchDir: filepath.Join(root, "data", "session-patches"),
	}

	if err := Sync(homebrewBin, paths, catalog, cfg, true); err != nil {
		t.Fatal(err)
	}

	// clother binary must NOT be copied into BinDir
	if _, err := os.Stat(filepath.Join(paths.BinDir, "clother.exe")); err == nil {
		t.Fatal("clother.exe must not be copied into BinDir in Homebrew mode")
	}

	// provider .bat files must exist and contain the Homebrew binary path
	for _, name := range []string{"claude", "clother-zai", "clother-native", "clother-or", "clother-custom"} {
		batPath := filepath.Join(paths.BinDir, name+".bat")
		content, err := os.ReadFile(batPath)
		if err != nil {
			t.Fatalf("missing %s.bat: %v", name, err)
		}
		if !strings.Contains(string(content), homebrewBin) {
			t.Fatalf("%s.bat should contain Homebrew binary path %q", name, homebrewBin)
		}
	}
}

func TestSyncHomebrewSkipsDynamicProviderLaunchers(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	homebrewBin := filepath.Join(root, "homebrew", "bin", "clother.exe")
	if err := os.MkdirAll(filepath.Dir(homebrewBin), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(homebrewBin, []byte("fake binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	catalog, err := providers.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.File{
		Version:           1,
		ProviderOverrides: map[string]config.ProviderOverride{},
		OpenRouterAliases: map[string]string{"kimi": "moonshotai/kimi-k2.5"},
		CustomProviders: map[string]config.CustomProvider{
			"myprovider": {Name: "myprovider", DisplayName: "myprovider", BaseURL: "https://example.com", APIKeyEnv: "MYPROVIDER_API_KEY"},
		},
	}
	paths := config.Paths{
		ConfigDir:       filepath.Join(root, "config"),
		DataDir:         filepath.Join(root, "data"),
		CacheDir:        filepath.Join(root, "cache"),
		BinDir:          filepath.Join(root, "bin"),
		ManifestFile:    filepath.Join(root, "data", "launchers.json"),
		SessionPatchDir: filepath.Join(root, "data", "session-patches"),
	}

	if err := Sync(homebrewBin, paths, catalog, cfg, true); err != nil {
		t.Fatal(err)
	}

	// individual dynamic .bat files must NOT be created under Homebrew
	for _, name := range []string{"clother-or-kimi", "clother-myprovider"} {
		batPath := filepath.Join(paths.BinDir, name+".bat")
		if _, err := os.Stat(batPath); err == nil {
			t.Fatalf("%s.bat must not be created in Homebrew mode", name)
		}
	}

	// gateway .bat files must always be present
	for _, name := range []string{"clother-or", "clother-custom"} {
		batPath := filepath.Join(paths.BinDir, name+".bat")
		if _, err := os.Stat(batPath); err != nil {
			t.Fatalf("gateway %s.bat must always be created: %v", name, err)
		}
	}
}
