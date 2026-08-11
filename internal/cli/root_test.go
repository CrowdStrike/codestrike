package cli

import "testing"

func TestNewRootCmd_HasExpectedSubcommands(t *testing.T) {
	root := NewRootCmd()

	if !root.SilenceErrors {
		t.Error("expected root command to leave error reporting to the caller")
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
	root.SetErr(nilWriter{})

	if err := root.Execute(); err == nil {
		t.Error("expected error when review is called without a PR URL, got nil")
	}
}

type nilWriter struct{}

func (nilWriter) Write(p []byte) (int, error) { return len(p), nil }
