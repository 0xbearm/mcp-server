package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	configEnvironmentVariable = "MCP_RUNNER_CONFIG"
	defaultConfigName         = "config.json"
)

type Config struct {
	Docker  string            `json:"docker"`
	Servers map[string]Server `json:"servers"`
}

type Server struct {
	Image      string            `json:"image"`
	Mounts     []Mount           `json:"mounts"`
	PassEnv    []string          `json:"pass_env"`
	Env        map[string]string `json:"env"`
	DockerArgs []string          `json:"docker_args"`
	Args       []string          `json:"args"`
}

type Mount struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	ReadOnly bool   `json:"read_only"`
}

func resolveConfigPath(override string) (string, error) {
	if override != "" {
		return expandHome(override)
	}
	if configured := os.Getenv(configEnvironmentVariable); configured != "" {
		return expandHome(configured)
	}
	configDirectory, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user configuration directory: %w", err)
	}
	return filepath.Join(configDirectory, "mcp-runner", defaultConfigName), nil
}

func loadConfig(filename string) (Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Config{}, fmt.Errorf("open config %q: %w", filename, err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return Config{}, fmt.Errorf("decode config %q: %w", filename, err)
	}
	if err := ensureJSONEnd(decoder); err != nil {
		return Config{}, fmt.Errorf("decode config %q: %w", filename, err)
	}

	if config.Docker == "" {
		config.Docker = "docker"
	}
	if len(config.Servers) == 0 {
		return Config{}, errors.New("config must define at least one server")
	}
	for name, server := range config.Servers {
		if strings.TrimSpace(name) == "" {
			return Config{}, errors.New("config contains an empty server name")
		}
		if strings.TrimSpace(server.Image) == "" {
			return Config{}, fmt.Errorf("server %q must define an image", name)
		}
	}
	return config, nil
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("multiple JSON values are not allowed")
	}
	return err
}

func expandHome(filename string) (string, error) {
	if filename == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("locate home directory: %w", err)
		}
		return home, nil
	}
	if strings.HasPrefix(filename, "~/") || strings.HasPrefix(filename, `~\`) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("locate home directory: %w", err)
		}
		return filepath.Join(home, filename[2:]), nil
	}
	if strings.HasPrefix(filename, "~") {
		return "", fmt.Errorf("user-specific home paths are not supported: %q", filename)
	}
	return filename, nil
}
