package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPrefersClaudeDirWhenNotOverridden(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	claudeDir := filepath.Join(root, "claude-bin")

	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// On Windows, look for claude.exe
	if err := os.WriteFile(filepath.Join(claudeDir, "claude.exe"), []byte("fake binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	t.Setenv("PATH", claudeDir)
	t.Setenv("CLOTHER_BIN", "")

	paths, err := Detect("")
	if err != nil {
		t.Fatal(err)
	}
	if paths.BinDir != claudeDir {
		t.Fatalf("Detect().BinDir = %q, want %q", paths.BinDir, claudeDir)
	}
}

func TestDetectPrefersCLOTHERBINOverClaudeDir(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	claudeDir := filepath.Join(root, "claude-bin")
	overrideDir := filepath.Join(root, "custom-bin")

	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// On Windows, look for claude.exe
	if err := os.WriteFile(filepath.Join(claudeDir, "claude.exe"), []byte("fake binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	t.Setenv("PATH", claudeDir)
	t.Setenv("CLOTHER_BIN", overrideDir)

	paths, err := Detect("")
	if err != nil {
		t.Fatal(err)
	}
	if paths.BinDir != overrideDir {
		t.Fatalf("Detect().BinDir = %q, want %q", paths.BinDir, overrideDir)
	}
}

func TestDetectPrefersExplicitFlagOverEnvAndClaudeDir(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	claudeDir := filepath.Join(root, "claude-bin")
	envDir := filepath.Join(root, "env-bin")
	flagDir := filepath.Join(root, "flag-bin")

	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// On Windows, look for claude.exe
	if err := os.WriteFile(filepath.Join(claudeDir, "claude.exe"), []byte("fake binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	t.Setenv("PATH", claudeDir)
	t.Setenv("CLOTHER_BIN", envDir)

	paths, err := Detect(flagDir)
	if err != nil {
		t.Fatal(err)
	}
	if paths.BinDir != flagDir {
		t.Fatalf("Detect().BinDir = %q, want %q", paths.BinDir, flagDir)
	}
}
