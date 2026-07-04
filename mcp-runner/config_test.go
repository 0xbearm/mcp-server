package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestFile(t *testing.T, filename, content string) {
	t.Helper()
	if err := os.WriteFile(filename, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestLoadConfig(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.json")
	content := `{
		"servers": {
			"python-ssh-mcp": {
				"image": "mcp-server:latest"
			}
		}
	}`
	writeTestFile(t, filename, content)

	config, err := loadConfig(filename)
	if err != nil {
		t.Fatal(err)
	}
	if config.Docker != "docker" {
		t.Fatalf("Docker = %q, want docker", config.Docker)
	}
	if config.Servers["python-ssh-mcp"].Image != "mcp-server:latest" {
		t.Fatal("server image was not decoded")
	}
}

func TestLoadConfigRejectsUnknownFields(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.json")
	content := `{
		"unknown": true,
		"servers": {
			"python-ssh-mcp": {
				"image": "mcp-server:latest"
			}
		}
	}`
	writeTestFile(t, filename, content)

	_, err := loadConfig(filename)
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}
