package commands

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jolehuit/clother/internal/config"
	"github.com/jolehuit/clother/internal/providers"
	"github.com/jolehuit/clother/internal/ui"
)

func TestRunInstallPreservesSameBinClaude(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	binDir := filepath.Join(root, "bin")

	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	t.Setenv("CLOTHER_BIN", binDir)
	t.Setenv("CLOTHER_SKIP_SELF_UPDATE", "1")

	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// On Windows, look for claude.exe
	realClaude := filepath.Join(binDir, "claude.exe")
	if err := os.WriteFile(realClaude, []byte("fake binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)

	paths, err := config.Detect("")
	if err != nil {
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
	output := &ui.Output{Stdout: io.Discard, Stderr: io.Discard, Format: ui.FormatHuman}

	code, err := runInstall(context.Background(), Context{
		Paths:   paths,
		Config:  cfg,
		Secrets: config.Secrets{},
		Catalog: catalog,
		Output:  output,
	})
	if err != nil {
		t.Fatalf("runInstall() error = %v", err)
	}
	if code != 0 {
		t.Fatalf("runInstall() code = %d, want 0", code)
	}

	// On Windows, we should have claude-real.exe
	if _, err := os.Stat(filepath.Join(binDir, "claude-real.exe")); err != nil {
		t.Fatalf("expected preserved real claude, stat error: %v", err)
	}
}

func TestRunInstallUpgradesToLatestRelease(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	binDir := filepath.Join(root, "bin")

	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	t.Setenv("CLOTHER_BIN", binDir)

	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// On Windows, look for claude.exe
	realClaude := filepath.Join(binDir, "claude.exe")
	if err := os.WriteFile(realClaude, []byte("fake binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)

	releaseBinary := filepath.Join(root, "release-clother.exe")
	if err := os.WriteFile(releaseBinary, []byte("release-3.0.3"), 0o755); err != nil {
		t.Fatal(err)
	}
	originalDownloader := downloadLatestBinary
	downloadLatestBinary = func(_ context.Context, _ string) (string, string, func(), error) {
		return releaseBinary, "v3.0.3", nil, nil
	}
	defer func() {
		downloadLatestBinary = originalDownloader
	}()

	paths, err := config.Detect("")
	if err != nil {
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

	code, err := runInstall(context.Background(), Context{
		Paths:   paths,
		Config:  cfg,
		Secrets: config.Secrets{},
		Catalog: catalog,
		Output:  &ui.Output{Stdout: io.Discard, Stderr: io.Discard, Format: ui.FormatHuman},
	})
	if err != nil {
		t.Fatalf("runInstall() error = %v", err)
	}
	if code != 0 {
		t.Fatalf("runInstall() code = %d, want 0", code)
	}

	installed, err := os.ReadFile(filepath.Join(binDir, "clother.exe"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(installed), "release-3.0.3") {
		t.Fatalf("expected installed clother to come from latest release, got %q", string(installed))
	}
}

func TestRunInstallWarnsWhenBinDirIsNotOnPath(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	binDir := filepath.Join(root, "bin")
	realClaudeDir := filepath.Join(root, "claude-bin")

	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	t.Setenv("CLOTHER_BIN", binDir)
	t.Setenv("CLOTHER_SKIP_SELF_UPDATE", "1")

	if err := os.MkdirAll(realClaudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// On Windows, create claude.exe
	if err := os.WriteFile(filepath.Join(realClaudeDir, "claude.exe"), []byte("fake binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", realClaudeDir+string(os.PathListSeparator)+oldPath)

	paths, err := config.Detect("")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := providers.Load()
	if err != nil {
		t.Fatal(err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	output := &ui.Output{Stdout: stdout, Stderr: stderr, Format: ui.FormatHuman}

	code, err := runInstall(context.Background(), Context{
		Paths:   paths,
		Config:  &config.File{Version: 1, ProviderOverrides: map[string]config.ProviderOverride{}, OpenRouterAliases: map[string]string{}, CustomProviders: map[string]config.CustomProvider{}},
		Secrets: config.Secrets{},
		Catalog: catalog,
		Output:  output,
	})
	if err != nil {
		t.Fatalf("runInstall() error = %v", err)
	}
	if code != 0 {
		t.Fatalf("runInstall() code = %d, want 0", code)
	}
	if !strings.Contains(stderr.String(), "is not on PATH") {
		t.Fatalf("expected PATH warning, got stderr %q", stderr.String())
	}
}
