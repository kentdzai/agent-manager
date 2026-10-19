//go:build windows

package agentsession

import (
	"os"
	"path/filepath"
)

// devinDataDir is Devin's per-user data directory on Windows:
// %APPDATA%\devin — the Rust dirs::data_dir layout, not %LOCALAPPDATA%.
func devinDataDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "devin")
}
