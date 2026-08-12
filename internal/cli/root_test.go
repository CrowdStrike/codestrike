package cli

import (
	"bytes"
	"testing"
)

func TestNewRootCmd_HasExpectedSubcommands(t *testing.T) {
	root := NewRootCmd()

	if root.SilenceErrors {
		t.Error("expected root command to report errors")
	}

	want := map[string]bool{"review": true, "version": true}
	got := map[string]bool{}
	for _, cmd := range root.Commands() {
		got[cmd.Name()] = true
	}

	for name := range want {
		if !got[name] {
			t.Errorf("expected root command to have subcommand %q, got %v", name, got)
		}
	}
}

func TestReviewCmd_RequiresExactlyOneArg(t *testing.T) {
	root := NewRootCmd()
	root.SetArgs([]string{"review"})
	root.SetOut(nilWriter{})
	var stderr bytes.Buffer
	root.SetErr(&stderr)

	if err := root.Execute(); err == nil {
		t.Error("expected error when review is called without a PR URL, got nil")
	}
	if stderr.Len() == 0 {
		t.Errorf("expected command to print returned error")
	}
}

type nilWriter struct{}

func (nilWriter) Write(p []byte) (int, error) { return len(p), nil }
