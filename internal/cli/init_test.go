package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCmd(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	configHome, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	root := NewRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"init"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), filepath.Join(configHome, "codestrike")) {
		t.Errorf("output = %q, want installed directory", out.String())
	}
	if _, err := os.Stat(filepath.Join(configHome, "codestrike", "default.yaml")); err != nil {
		t.Fatalf("default config was not installed: %v", err)
	}
}
