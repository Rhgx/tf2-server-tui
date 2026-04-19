package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ServerConfig struct {
	IP    string `json:"ip"`
	Port  int    `json:"port"`
	Label string `json:"label,omitempty"`
}

type Config struct {
	Servers         []ServerConfig `json:"servers"`
	RefreshInterval int            `json:"refreshInterval"`
}

var defaultConfig = Config{
	Servers: []ServerConfig{
		{IP: "127.0.0.1", Port: 27015, Label: "Example Server"},
	},
	RefreshInterval: 60,
}

// Load resolves the config path the app should use on startup. If explicitPath
// is set, Load uses that file directly. Otherwise it probes the working
// directory and executable directory, creates an example config on first run,
// and returns the resolved path alongside the parsed config.
func Load(explicitPath string) (Config, string, error) {
	if explicitPath != "" {
		cfg, err := Reload(explicitPath)
		return cfg, explicitPath, err
	}

	candidates := candidatePaths()
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			cfg, loadErr := loadFromPath(path)
			return cfg, path, loadErr
		}
	}

	targetPath := preferredCreatePath(candidates)
	if err := writeExampleConfig(targetPath); err != nil {
		return defaultConfig, targetPath, err
	}

	return defaultConfig, targetPath, fmt.Errorf(
		"created example config at %s; edit it and refresh",
		targetPath,
	)
}

// Reload re-reads the current config file. If the file is missing, Reload
// recreates the same example config that Load would create on first run so
// manual edits and in-app refreshes follow one consistent code path.
func Reload(path string) (Config, error) {
	if path == "" {
		return defaultConfig, errors.New("missing config path")
	}

	cfg, err := loadFromPath(path)
	if err == nil {
		return cfg, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		if writeErr := writeExampleConfig(path); writeErr != nil {
			return defaultConfig, writeErr
		}
		return defaultConfig, fmt.Errorf(
			"created example config at %s; edit it and refresh",
			path,
		)
	}

	return defaultConfig, err
}

// Save writes cfg back to path using the same JSON shape the app loads at
// startup, keeping the on-disk file simple for users to edit manually.
func Save(path string, cfg Config) error {
	if path == "" {
		return errors.New("missing config path")
	}

	if cfg.RefreshInterval <= 0 {
		cfg.RefreshInterval = defaultConfig.RefreshInterval
	}

	content, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	content = append(content, '\n')
	return os.WriteFile(path, content, 0o644)
}

// loadFromPath parses one JSON config file and drops obviously invalid server
// rows so a single bad entry does not make the whole config unusable.
func loadFromPath(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return defaultConfig, err
	}

	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return defaultConfig, fmt.Errorf("invalid JSON in %s: %w", path, err)
	}

	if cfg.RefreshInterval <= 0 {
		cfg.RefreshInterval = defaultConfig.RefreshInterval
	}

	validated := make([]ServerConfig, 0, len(cfg.Servers))
	for _, server := range cfg.Servers {
		server.IP = strings.TrimSpace(server.IP)
		if server.IP == "" || server.Port <= 0 || server.Port > 65535 {
			continue
		}
		validated = append(validated, server)
	}
	cfg.Servers = validated

	return cfg, nil
}

// candidatePaths returns the config locations the app should probe, preferring
// the current working directory and then the executable directory for portable
// builds.
func candidatePaths() []string {
	paths := make([]string, 0, 2)
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, "servers.json"))
	}
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		exeCandidate := filepath.Join(exeDir, "servers.json")
		if !containsPath(paths, exeCandidate) {
			paths = append(paths, exeCandidate)
		}
	}
	return paths
}

// preferredCreatePath picks the best place to create a new config file. It
// avoids writing into Go's temporary `go run` build directory, which would make
// the file hard for users to find.
func preferredCreatePath(candidates []string) string {
	for _, path := range candidates {
		if !looksLikeGoBuildTemp(path) {
			return path
		}
	}
	if len(candidates) > 0 {
		return candidates[0]
	}
	return "servers.json"
}

func looksLikeGoBuildTemp(path string) bool {
	lower := strings.ToLower(filepath.Clean(path))
	return strings.Contains(lower, `\appdata\local\temp\go-build`) ||
		strings.Contains(lower, `\temp\go-build`)
}

func containsPath(paths []string, want string) bool {
	for _, path := range paths {
		if samePath(path, want) {
			return true
		}
	}
	return false
}

func samePath(a string, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

func writeExampleConfig(path string) error {
	if path == "" {
		return errors.New("missing config path")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	content, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return err
	}

	content = append(content, '\n')
	return os.WriteFile(path, content, 0o644)
}
