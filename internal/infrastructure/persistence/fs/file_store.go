package fs

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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
		base = os.Getenv("SYZYGY_DATA_DIR")
		if base == "" {
			if home, err := os.UserHomeDir(); err == nil && home != "" {
				base = filepath.Join(home, ".syzygy-data")
			} else {
				base = "./syzygy-data"
			}
		}
	}
	return &FileStore{baseDir: base}
}

func (s *FileStore) GetOrCreateUnit(unitID string, title string, env map[string]any) (*domain.Unit, error) {
	u, err := s.GetUnit(unitID)
	if err == nil {
		if title != "" {
			u.Title = title
		}
		if env != nil {
			u.Env = env
		}
		u.UpdatedAt = time.Now().UTC()
		return u, s.SaveUnit(u)
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
	return u, s.SaveUnit(u)
}

func (s *FileStore) GetUnit(unitID string) (*domain.Unit, error) {
	path := s.unitPath(unitID)
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

func (s *FileStore) SaveUnit(u *domain.Unit) error {
	if err := os.MkdirAll(filepath.Dir(s.unitPath(u.UnitID)), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.unitPath(u.UnitID), b, 0o644)
}

func (s *FileStore) unitPath(unitID string) string {
	return filepath.Join(s.baseDir, "units", unitID+".json")
}
