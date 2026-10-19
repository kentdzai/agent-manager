//go:build !windows

package agentsession

import (
	"path/filepath"
	"runtime"
)

// devinDataDir is Devin's per-user data directory: $XDG_DATA_HOME/devin on
// Linux (or ~/.local/share/devin), ~/Library/Application Support/devin on
// macOS — the layout the Rust dirs::data_dir convention gives the CLI.
func devinDataDir(env Env) string {
	home, err := env.homeDir()
	if err != nil {
		return ""
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "devin")
	}
	if dir := env.get("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, "devin")
	}
	return filepath.Join(home, ".local", "share", "devin")
}
