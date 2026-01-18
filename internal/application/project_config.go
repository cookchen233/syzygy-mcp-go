package application

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ProjectConfig struct {
	ProjectKey    string            `json:"project_key"`
	Env           map[string]string `json:"env"`
	RunnerCommand string            `json:"runner_command"`
	RunnerDir     string            `json:"runner_dir"`
	ArtifactsDir  string            `json:"artifacts_dir"`
	UpdatedAt     string            `json:"updated_at"`
}

func (s *SyzygyService) projectConfigPath(projectKey string) (string, error) {
	base := s.store.BaseDir()
	if base == "" {
		base = os.Getenv("SYZYGY_HOME")
	}
	if base == "" {
		return "", fmt.Errorf("SYZYGY_HOME is empty")
	}
	return filepath.Join(base, "projects", safeProjectKey(projectKey), "config.json"), nil
}

func (s *SyzygyService) LoadProjectConfig(projectKey string) (*ProjectConfig, error) {
	p, err := s.projectConfigPath(projectKey)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var cfg ProjectConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	if cfg.Env == nil {
		cfg.Env = map[string]string{}
	}
	return &cfg, nil
}

func (s *SyzygyService) SaveProjectConfig(cfg *ProjectConfig) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("nil config")
	}
	p, err := s.projectConfigPath(cfg.ProjectKey)
	if err != nil {
		return "", err
	}
	if cfg.Env == nil {
		cfg.Env = map[string]string{}
	}
	cfg.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(p, b, 0o644); err != nil {
		return "", err
	}
	return p, nil
}

func (s *SyzygyService) EnsureProjectInitialized(projectKey string) (*ProjectConfig, error) {
	cfg, err := s.LoadProjectConfig(projectKey)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewAppError("project_not_initialized", "project config not found; call syzygy_project_init first")
		}
		return nil, err
	}
	return cfg, nil
}

func safeProjectKey(projectKey string) string {
	k := strings.TrimSpace(projectKey)
	if k == "" {
		return "default"
	}
	k = strings.ReplaceAll(k, "..", "")
	k = strings.ReplaceAll(k, string(filepath.Separator), "-")
	k = strings.ReplaceAll(k, "/", "-")
	return k
}

func anyToString(v any) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func normalizeRunnerCommand(s0 string) string {
	s0 = strings.TrimSpace(s0)
	if s0 == "" {
		return ""
	}
	return s0
}
