package config

import (
	"os"
	"os/exec"
	"path/filepath"
)

type Paths struct {
	ConfigDir       string
	DataDir         string
	CacheDir        string
	BinDir          string
	ConfigFile      string
	SecretsFile     string
	ManifestFile    string
	SessionPatchDir string
	UpdateCacheFile string
}

func Detect(binOverride string) (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}

	// Windows paths: %APPDATA% for config, %LOCALAPPDATA% for data/cache
	appData := getenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	localAppData := getenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))

	configDir := getenv("CLOTHER_CONFIG_DIR", filepath.Join(appData, "clother"))
	dataDir := getenv("CLOTHER_DATA_DIR", filepath.Join(localAppData, "clother"))
	cacheDir := getenv("CLOTHER_CACHE_DIR", filepath.Join(localAppData, "clother", "cache"))

	binDir := getenv("CLOTHER_BIN", "")
	if binOverride != "" {
		binDir = binOverride
	}
	if binDir == "" {
		binDir = defaultBinDir(home)
	}

	return Paths{
		ConfigDir:       configDir,
		DataDir:         dataDir,
		CacheDir:        cacheDir,
		BinDir:          binDir,
		ConfigFile:      filepath.Join(configDir, "config.json"),
		SecretsFile:     filepath.Join(dataDir, "secrets.env"),
		ManifestFile:    filepath.Join(dataDir, "launchers.json"),
		SessionPatchDir: filepath.Join(dataDir, "session-patches"),
		UpdateCacheFile: filepath.Join(cacheDir, "update.json"),
	}, nil
}

func (p Paths) EnsureBaseDirs() error {
	for _, dir := range []string{p.ConfigDir, p.DataDir, p.CacheDir, p.SessionPatchDir, p.BinDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func defaultBinDir(home string) string {
	if dir := claudeBinDir(); dir != "" {
		return dir
	}
	// Windows: use %LOCALAPPDATA%\clother\bin
	localAppData := getenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	return filepath.Join(localAppData, "clother", "bin")
}

func claudeBinDir() string {
	claudePath, err := exec.LookPath("claude")
	if err != nil || claudePath == "" {
		return ""
	}
	if abs, err := filepath.Abs(claudePath); err == nil {
		claudePath = abs
	}
	return filepath.Dir(claudePath)
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
