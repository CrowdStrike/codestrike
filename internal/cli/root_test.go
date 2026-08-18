package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewRootCmd_HasExpectedSubcommands(t *testing.T) {
	root := NewRootCmd()

	if root.SilenceErrors {
		t.Error("expected root command to report errors")
	}

	want := map[string]bool{"init": true, "review": true, "version": true}
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

func TestRootCmd_HasConfigFlag(t *testing.T) {
	root := NewRootCmd()
	flag := root.PersistentFlags().Lookup("config")
	if flag == nil {
		t.Fatal("expected root command to have a --config persistent flag")
	}
	if flag.DefValue != "" {
		t.Errorf("--config default = %q, want empty string", flag.DefValue)
	}
}

func TestReviewCmd_ConfigFlagNotFound(t *testing.T) {
	root := NewRootCmd()
	root.SetArgs([]string{
		"review",
		"--config", "/nonexistent/codestrike.yaml",
		"https://github.com/owner/repo/pull/1",
	})
	root.SetOut(nilWriter{})
	var stderr bytes.Buffer
	root.SetErr(&stderr)

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing --config file")
	}
	if !strings.Contains(err.Error(), "/nonexistent/codestrike.yaml") {
		t.Errorf("error = %q, want it to mention the --config path", err.Error())
	}
}
