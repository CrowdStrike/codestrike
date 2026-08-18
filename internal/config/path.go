package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	appConfigDirName  = "codestrike"
	appConfigFileName = "default.yaml"
)

// ResolvePath returns the config file path to use: flagValue if non-empty,
// otherwise the OS default user config location (codestrike/default.yaml
// under os.UserConfigDir()).
func ResolvePath(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("determining user config directory: %w", err)
	}

	return filepath.Join(dir, appConfigDirName, appConfigFileName), nil
}
