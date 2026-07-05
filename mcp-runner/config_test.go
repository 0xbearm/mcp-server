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
			"python-ssh-mcp": {}
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
	if config.Image != defaultImage {
		t.Fatalf("Image = %q, want %q", config.Image, defaultImage)
	}
	if _, ok := config.Servers["python-ssh-mcp"]; !ok {
		t.Fatal("server was not decoded")
	}
}

func TestLoadConfigRejectsUnknownFields(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.json")
	content := `{
		"unknown": true,
		"servers": {
			"python-ssh-mcp": {}
		}
	}`
	writeTestFile(t, filename, content)

	_, err := loadConfig(filename)
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}

func TestLoadConfigUsesTopLevelImage(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.json")
	writeTestFile(t, filename, `{
		"image": "registry.example/mcp-server:test",
		"servers": {
			"example": {}
		}
	}`)

	config, err := loadConfig(filename)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := config.Image, "registry.example/mcp-server:test"; got != want {
		t.Fatalf("Image = %q, want %q", got, want)
	}
}
