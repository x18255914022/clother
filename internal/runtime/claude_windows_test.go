package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jolehuit/clother/internal/config"
)

func TestFindRealClaudeCanUseSameBinDir(t *testing.T) {
	root := t.TempDir()
	binDir := filepath.Join(root, "clother-bin")
	realDir := filepath.Join(root, "real-bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// On Windows, create claude.exe
	if err := os.WriteFile(filepath.Join(binDir, "claude.exe"), []byte("fake binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	realClaude := filepath.Join(realDir, "claude.exe")
	if err := os.WriteFile(realClaude, []byte("fake binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	oldPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", oldPath) })
	if err := os.Setenv("PATH", binDir+string(os.PathListSeparator)+realDir); err != nil {
		t.Fatal(err)
	}

	got, err := FindRealClaude(config.Paths{BinDir: binDir})
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join(binDir, "claude.exe") {
		t.Fatalf("FindRealClaude() = %q, want %q", got, filepath.Join(binDir, "claude.exe"))
	}
}

func TestFindRealClaudeSkipsSelfAndFallsBack(t *testing.T) {
	// On Windows, symlinks require elevated privileges, so we test the fallback
	// path by ensuring that when no claude is in PATH, the fallback is used.
	root := t.TempDir()
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	realFallback := filepath.Join(binDir, "claude-real.exe")
	if err := os.WriteFile(realFallback, []byte("fake binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Set PATH to empty so no claude is found
	oldPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", oldPath) })
	if err := os.Setenv("PATH", ""); err != nil {
		t.Fatal(err)
	}

	got, err := FindRealClaude(config.Paths{BinDir: binDir})
	if err != nil {
		t.Fatal(err)
	}
	if got != realFallback {
		t.Fatalf("FindRealClaude() = %q, want %q", got, realFallback)
	}
}

func TestPreserveRealClaudeMovesClaudeToClaudeReal(t *testing.T) {
	root := t.TempDir()
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	claudePath := filepath.Join(binDir, "claude.exe")
	content := []byte("real-claude-binary")
	if err := os.WriteFile(claudePath, content, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := PreserveRealClaude(config.Paths{BinDir: binDir}, claudePath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(claudePath); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be moved, stat err=%v", claudePath, err)
	}
	preserved := filepath.Join(binDir, "claude-real.exe")
	got, err := os.ReadFile(preserved)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Fatalf("preserved content mismatch: got %q", string(got))
	}
}
