package launchers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jolehuit/clother/internal/config"
	"github.com/jolehuit/clother/internal/profiles"
	"github.com/jolehuit/clother/internal/providers"
)

type Manifest struct {
	Launchers []string `json:"launchers"`
}

// Sync installs the clother binary and provider batch wrappers into paths.BinDir.
//
// On Windows, we create .bat wrapper scripts instead of symlinks because
// symlinks require elevated privileges or Developer Mode.
func Sync(execPath string, paths config.Paths, catalog providers.Catalog, cfg *config.File, skipCopy bool) error {
	if err := paths.EnsureBaseDirs(); err != nil {
		return err
	}

	if !skipCopy {
		destBinary := filepath.Join(paths.BinDir, "clother.exe")
		if err := copyExecutable(execPath, destBinary); err != nil {
			return err
		}
	}

	previous, _ := LoadManifest(paths.ManifestFile)
	desired := map[string]struct{}{}
	for _, target := range profiles.All(catalog, cfg) {
		if skipCopy && isDynamicProfile(target.Profile, cfg) {
			continue
		}
		desired[launcherName(target.Profile)] = struct{}{}
	}
	desired["clother-or"] = struct{}{}
	desired["clother-custom"] = struct{}{}

	for _, old := range previous.Launchers {
		if _, ok := desired[old]; ok {
			continue
		}
		// Remove old .bat and .exe files
		_ = os.Remove(filepath.Join(paths.BinDir, old+".bat"))
		_ = os.Remove(filepath.Join(paths.BinDir, old+".exe"))
		_ = os.Remove(filepath.Join(paths.BinDir, old))
	}

	var launchers []string
	for name := range desired {
		launchers = append(launchers, name)
	}
	sort.Strings(launchers)
	for _, name := range launchers {
		if err := createBatchWrapper(paths.BinDir, name, skipCopy, execPath); err != nil {
			return err
		}
	}
	// Create claude shim
	if err := createBatchWrapper(paths.BinDir, "claude", skipCopy, execPath); err != nil {
		return err
	}

	return SaveManifest(paths.ManifestFile, Manifest{Launchers: launchers})
}

// createBatchWrapper creates a .bat file that invokes clother with the correct profile.
func createBatchWrapper(binDir, name string, skipCopy bool, execPath string) error {
	var target string
	if skipCopy {
		target = execPath
	} else {
		target = filepath.Join(binDir, "clother.exe")
	}

	batContent := "@echo off\r\n\"" + target + "\" %*\r\n"

	batPath := filepath.Join(binDir, name+".bat")
	return writeAtomic(batPath, []byte(batContent), 0o755)
}

func LoadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func SaveManifest(path string, manifest Manifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeAtomic(path, data, 0o644)
}

func launcherName(profile string) string {
	return "clother-" + profile
}

func isDynamicProfile(profile string, cfg *config.File) bool {
	if strings.HasPrefix(profile, "or-") {
		return true
	}
	if cfg == nil {
		return false
	}
	_, isCustom := cfg.CustomProviders[profile]
	return isCustom
}

func copyExecutable(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return writeAtomic(dst, data, 0o755)
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".launcher-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
