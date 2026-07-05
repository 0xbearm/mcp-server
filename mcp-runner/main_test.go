package main

import (
	"bytes"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestHelpSucceeds(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := run([]string{"--help"}, &stdout, &stderr); err != nil {
		t.Fatalf("run --help: %v", err)
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("help output did not contain usage: %q", stderr.String())
	}
}

func TestDryRun(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "config.json")
	writeTestFile(t, configFile, `{
		"servers": {
			"example": {}
		}
	}`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(
		[]string{"--config", configFile, "--dry-run", "example"},
		&stdout,
		&stderr,
	)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(stdout.String())
	want := regexp.MustCompile(`^docker run --rm -i --name example-[0-9a-f]{16} mcp-server:latest example$`)
	if !want.MatchString(got) {
		t.Fatalf("dry-run output = %q, want output matching %q", got, want)
	}
}

func TestDebugPrintsDockerCommandToStderrBeforeExecution(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "config.json")
	writeTestFile(t, configFile, `{
		"docker": "mcp-runner-test-nonexistent-docker",
		"servers": {
			"example": {}
		}
	}`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(
		[]string{"--config", configFile, "--debug", "example"},
		&stdout,
		&stderr,
	)
	if err == nil {
		t.Fatal("expected Docker execution to fail")
	}
	if stdout.Len() != 0 {
		t.Fatalf("debug wrote to stdout: %q", stdout.String())
	}
	got := strings.TrimSpace(stderr.String())
	want := regexp.MustCompile(`^mcp-runner-test-nonexistent-docker run --rm -i --name example-[0-9a-f]{16} mcp-server:latest example$`)
	if !want.MatchString(got) {
		t.Fatalf("debug output = %q, want output matching %q", got, want)
	}
}
