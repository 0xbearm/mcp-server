package main

import (
	"bytes"
	"path/filepath"
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
			"example": {
				"image": "example:latest"
			}
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
	if got, want := strings.TrimSpace(stdout.String()), "docker run --rm -i example:latest example"; got != want {
		t.Fatalf("dry-run output = %q, want %q", got, want)
	}
}
