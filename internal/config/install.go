package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Install writes the bundled default configuration to the OS user config
// directory. Existing files are preserved unless force is true.
func Install(force bool) (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("determining user config directory: %w", err)
	}
	targetDir := filepath.Join(userConfigDir, appConfigDirName)
	assets, err := fs.Sub(embeddedAssets, "assets")
	if err != nil {
		return "", fmt.Errorf("opening embedded config files: %w", err)
	}

	if !force {
		var existing string
		err := fs.WalkDir(assets, ".", func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil || entry.IsDir() || existing != "" {
				return walkErr
			}
			target := filepath.Join(targetDir, filepath.FromSlash(path))
			if _, statErr := os.Stat(target); statErr == nil {
				existing = target
			} else if !errors.Is(statErr, fs.ErrNotExist) {
				return statErr
			}
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("checking existing config files: %w", err)
		}
		if existing != "" {
			return "", fmt.Errorf("config file %q already exists; use --force to overwrite bundled files", existing)
		}
	}

	err = fs.WalkDir(assets, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		target := filepath.Join(targetDir, filepath.FromSlash(path))
		if entry.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, readErr := fs.ReadFile(assets, path)
		if readErr != nil {
			return readErr
		}
		return os.WriteFile(target, data, 0600)
	})
	if err != nil {
		return "", fmt.Errorf("writing config files: %w", err)
	}

	return targetDir, nil
}
