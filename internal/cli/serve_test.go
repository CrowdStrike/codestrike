package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestServeCmd_IsRegistered(t *testing.T) {
	root := NewRootCmd()

	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "serve" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected root command to have a 'serve' subcommand")
	}
}

func TestServeCmd_HasAddrFlag(t *testing.T) {
	root := NewRootCmd()

	var serve *cobra.Command
	for _, cmd := range root.Commands() {
		if cmd.Name() == "serve" {
			serve = cmd
			break
		}
	}

	if serve == nil {
		t.Fatal("serve subcommand not found")
	}

	flag := serve.Flags().Lookup("addr")
	if flag == nil {
		t.Fatal("expected serve command to have an --addr flag")
	}
	if flag.DefValue != "" {
		t.Errorf("--addr default = %q, want empty string", flag.DefValue)
	}
}

func TestServeCmd_RejectsArgs(t *testing.T) {
	root := NewRootCmd()
	root.SetArgs([]string{"serve", "unexpected-arg"})
	root.SetOut(nilWriter{})
	var stderr bytes.Buffer
	root.SetErr(&stderr)

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when serve is called with args")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("error = %q, want it to mention unknown command", err.Error())
	}
}

func TestServeCmd_ConfigFlagNotFound(t *testing.T) {
	root := NewRootCmd()
	root.SetArgs([]string{
		"serve",
		"--config", "/nonexistent/codestrike.yaml",
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
