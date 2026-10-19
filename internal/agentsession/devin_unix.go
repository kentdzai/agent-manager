//go:build !windows

package agentsession

import (
	"os"
	"path/filepath"
	"runtime"
)

// devinDataDir is Devin's per-user data directory: $XDG_DATA_HOME/devin on
// Linux (or ~/.local/share/devin), ~/Library/Application Support/devin on
// macOS — the layout the Rust dirs::data_dir convention gives the CLI.
func devinDataDir() string {
	if runtime.GOOS == "darwin" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return ""
		}
		return filepath.Join(dir, "devin")
	}
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, "devin")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "share", "devin")
}
