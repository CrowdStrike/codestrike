package context_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	appcontext "github.com/CrowdStrike/codestrike/internal/context"
)

func TestDiscoverCursorContext_ReadsAgentsMdAndAlwaysApplyRules(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "AGENTS.md", "# Project Instructions\nUse snake_case.\n")

	rulesDir := filepath.Join(dir, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, rulesDir, "always.mdc", "---\nalwaysApply: true\n---\n\nAlways follow this rule.\n")
	writeFile(t, rulesDir, "no-frontmatter-fields.mdc", "---\ndescription: agent decides\n---\n\nAgent-decided rule body.\n")

	got := appcontext.DiscoverCursorContext(dir)

	if !strings.Contains(got, "### AGENTS.md") || !strings.Contains(got, "Use snake_case.") {
		t.Errorf("expected AGENTS.md content in output, got: %q", got)
	}
	if !strings.Contains(got, "### .cursor/rules/always.mdc") || !strings.Contains(got, "Always follow this rule.") {
		t.Errorf("expected always.mdc content in output, got: %q", got)
	}
	if !strings.Contains(got, "### .cursor/rules/no-frontmatter-fields.mdc") || !strings.Contains(got, "Agent-decided rule body.") {
		t.Errorf("expected no-frontmatter-fields.mdc content in output, got: %q", got)
	}
}

func TestDiscoverCursorContext_SkipsGlobScopedAndDisabledRules(t *testing.T) {
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, rulesDir, "scoped.mdc", "---\nglobs: \"**/*.tsx\"\nalwaysApply: false\n---\n\nOnly for tsx files.\n")
	writeFile(t, rulesDir, "disabled.mdc", "---\nalwaysApply: false\n---\n\nManual only.\n")
	writeFile(t, rulesDir, "not-a-rule.md", "not an mdc file\n")

	got := appcontext.DiscoverCursorContext(dir)

	if strings.Contains(got, "scoped.mdc") || strings.Contains(got, "Only for tsx files.") {
		t.Errorf("expected glob-scoped rule to be excluded, got: %q", got)
	}
	if strings.Contains(got, "disabled.mdc") || strings.Contains(got, "Manual only.") {
		t.Errorf("expected alwaysApply:false rule to be excluded, got: %q", got)
	}
	if strings.Contains(got, "not-a-rule") {
		t.Errorf("expected non-.mdc file to be ignored, got: %q", got)
	}
}

func TestDiscoverCursorContext_MissingDirReturnsEmptyNoError(t *testing.T) {
	dir := t.TempDir()

	got := appcontext.DiscoverCursorContext(dir)

	if got != "" {
		t.Errorf("expected empty result for repo with no AGENTS.md/.cursor/rules, got: %q", got)
	}
}

func TestDiscoverCursorContext_SkipsFileWithoutFrontmatter(t *testing.T) {
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, rulesDir, "plain.mdc", "Just plain text, no frontmatter delimiters.\n")

	got := appcontext.DiscoverCursorContext(dir)

	if got != "" {
		t.Errorf("expected file without frontmatter to be skipped, got: %q", got)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}
