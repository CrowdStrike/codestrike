package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// cursorRuleFrontmatter mirrors the subset of Cursor's .mdc frontmatter
// fields relevant to deciding whether a rule is unconditionally applicable
// in a whole-PR review, where there is no single-file diff-aware matching
// and no interactive chat session for @-mentions to happen in.
type cursorRuleFrontmatter struct {
	Description string `yaml:"description"`
	Globs       string `yaml:"globs"`
	AlwaysApply *bool  `yaml:"alwaysApply"`
}

// parseMDCFrontmatter splits an .mdc file into its frontmatter and body. ok
// is false when the file has no valid "---" delimited frontmatter block, in
// which case the caller should skip it.
func parseMDCFrontmatter(data []byte) (fm cursorRuleFrontmatter, body string, ok bool) {
	s := string(data)
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		return cursorRuleFrontmatter{}, "", false
	}

	rest := s[strings.IndexByte(s, '\n')+1:]
	end := strings.Index(rest, "\n---\n")
	sepLen := len("\n---\n")
	if end == -1 {
		end = strings.Index(rest, "\n---\r\n")
		sepLen = len("\n---\r\n")
	}
	if end == -1 {
		return cursorRuleFrontmatter{}, "", false
	}

	raw := rest[:end]
	remainder := rest[end+sepLen:]

	if err := yaml.Unmarshal([]byte(raw), &fm); err != nil {
		return cursorRuleFrontmatter{}, "", false
	}
	return fm, strings.TrimSpace(remainder), true
}

// isUnconditionallyApplicable reports whether a rule should be folded into
// a whole-PR review's project context: alwaysApply: true, or rules with no
// file-scoping (globs) at all. Rules scoped to specific files via globs are
// skipped, since a whole-PR review has no single file to match them against.
func isUnconditionallyApplicable(fm cursorRuleFrontmatter) bool {
	if fm.AlwaysApply != nil {
		return *fm.AlwaysApply
	}
	return strings.TrimSpace(fm.Globs) == ""
}

// DiscoverCursorContext reads AGENTS.md and applicable .cursor/rules/*.mdc
// files from repoRoot (root level only, not recursive) and returns their
// concatenated content formatted as "### <name>" sections, consistent with
// how project context_files are rendered.
func DiscoverCursorContext(repoRoot string) string {
	var sb strings.Builder

	if data, err := os.ReadFile(filepath.Join(repoRoot, "AGENTS.md")); err == nil {
		if content := strings.TrimSpace(string(data)); content != "" {
			sb.WriteString("### AGENTS.md\n")
			sb.WriteString(content)
			sb.WriteString("\n\n")
		}
	}

	rulesDir := filepath.Join(repoRoot, ".cursor", "rules")
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return sb.String()
	}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".mdc" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(rulesDir, e.Name()))
		if err != nil {
			continue
		}
		fm, body, ok := parseMDCFrontmatter(data)
		if !ok || body == "" || !isUnconditionallyApplicable(fm) {
			continue
		}
		fmt.Fprintf(&sb, "### .cursor/rules/%s\n%s\n\n", e.Name(), body)
	}

	return sb.String()
}
