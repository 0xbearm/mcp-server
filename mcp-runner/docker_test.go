package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBuildDockerArgs(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configFile, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SSH_PASSWORD", "secret")

	server := Server{
		Image: "mcp-server:latest",
		Mounts: []Mount{{
			Source:   configFile,
			Target:   "/root/.config/python-ssh-mcp/config.json",
			ReadOnly: true,
		}},
		PassEnv: []string{"SSH_PASSWORD"},
		Env:     map[string]string{"LOG_LEVEL": "info"},
		Args:    []string{"--configured"},
	}

	got, err := buildDockerArgs("python-ssh-mcp", server, []string{"--verbose"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"run", "--rm", "-i",
		"--mount", "type=bind,src=" + configFile + ",dst=/root/.config/python-ssh-mcp/config.json,readonly",
		"--env", "LOG_LEVEL=info",
		"--env", "SSH_PASSWORD",
		"mcp-server:latest",
		"python-ssh-mcp",
		"--configured",
		"--verbose",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("arguments differ\n got: %#v\nwant: %#v", got, want)
	}
}

func TestBuildDockerArgsRejectsTTY(t *testing.T) {
	_, err := buildDockerArgs("example", Server{
		Image:      "example",
		DockerArgs: []string{"--tty"},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "TTY") {
		t.Fatalf("expected TTY error, got %v", err)
	}
}

func TestBuildDockerArgsRequiresPassedEnvironment(t *testing.T) {
	const name = "MCP_RUNNER_TEST_UNSET"
	t.Setenv(name, "temporary")
	if err := os.Unsetenv(name); err != nil {
		t.Fatal(err)
	}

	_, err := buildDockerArgs("example", Server{
		Image:   "example",
		PassEnv: []string{name},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "is not set") {
		t.Fatalf("expected missing environment error, got %v", err)
	}
}

func TestShellQuote(t *testing.T) {
	tests := map[string]string{
		"simple":       "simple",
		"with spaces":  "'with spaces'",
		"it's":         `'it'"'"'s'`,
		"":             "''",
		"/tmp/a:b,c.d": "/tmp/a:b,c.d",
	}
	for input, want := range tests {
		if got := shellQuote(input); got != want {
			t.Errorf("shellQuote(%q) = %q, want %q", input, got, want)
		}
	}
}
