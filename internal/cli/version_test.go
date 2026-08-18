package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCmd_Short(t *testing.T) {
	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(out.String())
	if !strings.HasPrefix(got, "codestrike ") {
		t.Errorf("expected output to start with %q, got %q", "codestrike ", got)
	}
}

func TestVersionCmd_Long(t *testing.T) {
	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"version", "--long"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	for _, want := range []string{"Version:", "Commit:", "Commit Date:", "Built By:", "Build Date:"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected output to contain %q, got %q", want, got)
		}
	}
}
