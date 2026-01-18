package fs

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cookchen233/syzygy-mcp-go/internal/domain"
)

type FileStoreConfig struct {
	BaseDir string
}

type FileStore struct {
	baseDir string
}

func (s *FileStore) BaseDir() string {
	return s.baseDir
}

func NewFileStore(cfg FileStoreConfig) *FileStore {
	base := cfg.BaseDir
	if base == "" {
		base = os.Getenv("SYZYGY_HOME")
		if base == "" {
			if home, err := os.UserHomeDir(); err == nil && home != "" {
				base = filepath.Join(home, ".syzygy-mcp")
			} else {
				base = "./.syzygy-mcp"
			}
		}
	}
	return &FileStore{baseDir: base}
}

func (s *FileStore) GetOrCreateUnit(projectKey string, unitID string, title string, env map[string]any) (*domain.Unit, error) {
	u, err := s.GetUnit(projectKey, unitID)
	if err == nil {
		if title != "" {
			u.Title = title
		}
		if env != nil {
			u.Env = env
		}
		u.UpdatedAt = time.Now().UTC()
		return u, s.SaveUnit(projectKey, u)
	}

	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	now := time.Now().UTC()
	u = &domain.Unit{
		UnitID:    unitID,
		Title:     title,
		Env:       env,
		Runs:      []*domain.Run{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	return u, s.SaveUnit(projectKey, u)
}

func (s *FileStore) GetUnit(projectKey string, unitID string) (*domain.Unit, error) {
	path := s.unitPath(projectKey, unitID)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var u domain.Unit
	if err := json.Unmarshal(b, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *FileStore) SaveUnit(projectKey string, u *domain.Unit) error {
	if err := os.MkdirAll(filepath.Dir(s.unitPath(projectKey, u.UnitID)), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.unitPath(projectKey, u.UnitID), b, 0o644)
}

func (s *FileStore) ListUnitIDs(projectKey string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(s.baseDir, "projects", safeProjectKey(projectKey), "units"))
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	ids := []string{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		ids = append(ids, strings.TrimSuffix(name, ".json"))
	}
	return ids, nil
}

func (s *FileStore) unitPath(projectKey string, unitID string) string {
	return filepath.Join(s.baseDir, "projects", safeProjectKey(projectKey), "units", unitID+".json")
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
