package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CrowdStrike/codestrike/internal/config"
)

func TestInstall(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	configHome, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}

	dir, err := config.Install(false)
	if err != nil {
		t.Fatal(err)
	}
	wantDir := filepath.Join(configHome, "codestrike")
	if dir != wantDir {
		t.Fatalf("Install() dir = %q, want %q", dir, wantDir)
	}

	for _, name := range []string{
		"default.yaml",
		"prompts/default.md",
		"tones/default.md",
		"tones/direct.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected installed file %q: %v", name, err)
		}
	}
}

func TestInstall_DoesNotOverwriteWithoutForce(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir, err := config.Install(false)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "default.yaml")
	if err := os.WriteFile(path, []byte("custom"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := config.Install(false); err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("Install() error = %v, want --force guidance", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "custom" {
		t.Fatalf("existing config was overwritten: %q", data)
	}

	if _, err := config.Install(true); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "custom" {
		t.Fatal("--force did not overwrite existing config")
	}
}
