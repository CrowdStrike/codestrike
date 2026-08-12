package config_test

import (
	"path/filepath"
	"testing"

	"github.com/CrowdStrike/codestrike/internal/config"
)

func TestResolvePath_ExplicitFlagWins(t *testing.T) {
	got, err := config.ResolvePath("/custom/path.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/custom/path.yaml" {
		t.Errorf("ResolvePath = %q, want %q", got, "/custom/path.yaml")
	}
}

func TestResolvePath_DefaultUsesOSUserConfigDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	got, err := config.ResolvePath("")
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Base(got) != "default.yaml" {
		t.Errorf("ResolvePath() filename = %q, want %q", filepath.Base(got), "default.yaml")
	}
	if filepath.Base(filepath.Dir(got)) != "codestrike" {
		t.Errorf("ResolvePath() parent dir = %q, want %q", filepath.Base(filepath.Dir(got)), "codestrike")
	}
	if !filepath.IsAbs(got) {
		t.Errorf("ResolvePath() = %q, want absolute path", got)
	}
}
