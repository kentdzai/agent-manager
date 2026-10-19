//go:build windows

package agentsession

import (
	"os"
	"path/filepath"
)

// devinDataDir is Devin's per-user data directory on Windows:
// %APPDATA%\devin — the Rust dirs::data_dir layout, not %LOCALAPPDATA%.
func devinDataDir(env Env) string {
	dir := env.get("APPDATA")
	if dir == "" {
		var err error
		if dir, err = os.UserConfigDir(); err != nil {
			return ""
		}
	}
	return filepath.Join(dir, "devin")
}
