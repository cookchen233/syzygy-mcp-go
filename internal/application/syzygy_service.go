package application

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cookchen233/syzygy-mcp-go/internal/domain"
)

type SyzygyService struct {
	store  Store
	logger *log.Logger
}

func NewSyzygyService(store Store, logger *log.Logger) *SyzygyService {
	if logger == nil {
		logger = log.Default()
	}
	return &SyzygyService{store: store, logger: logger}
}

func (s *SyzygyService) UnitStart(unitID, title string, env map[string]any, variables map[string]any) (map[string]any, error) {
	u, err := s.store.GetOrCreateUnit(unitID, title, env)
	if err != nil {
		return nil, err
	}

	runID, err := domain.NewID("run")
	if err != nil {
		return nil, err
	}

	run := &domain.Run{
		RunID:     runID,
		Status:    "in_progress",
		Variables: variables,
		Steps:     []*domain.ActionStep{},
		Anchors:   map[string]string{},
		DBChecks:  []*domain.DbCheck{},
		Artifacts: map[string]string{},
		StartedAt: time.Now().UTC(),
		Meta:      map[string]any{},
	}

	u.Runs = append(u.Runs, run)
	u.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveUnit(u); err != nil {
		return nil, err
	}

	return map[string]any{"unit_id": unitID, "run_id": runID}, nil
}

func (s *SyzygyService) StepAppend(unitID, runID string, step domain.ActionStep) (map[string]any, error) {
	u, err := s.store.GetUnit(unitID)
	if err != nil {
		return nil, err
	}

	run, err := findRun(u, runID)
	if err != nil {
		return nil, err
	}

	stepID, err := domain.NewID("step")
	if err != nil {
		return nil, err
	}
	step.StepID = stepID
	run.Steps = append(run.Steps, &step)
	u.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveUnit(u); err != nil {
		return nil, err
	}

	return map[string]any{"step_id": stepID}, nil
}

func (s *SyzygyService) AnchorSet(unitID, runID, key, value, source string) (map[string]any, error) {
	u, err := s.store.GetUnit(unitID)
	if err != nil {
		return nil, err
	}
	run, err := findRun(u, runID)
	if err != nil {
		return nil, err
	}
	if run.Anchors == nil {
		run.Anchors = map[string]string{}
	}
	run.Anchors[key] = value
	if run.Meta == nil {
		run.Meta = map[string]any{}
	}
	run.Meta["last_anchor_source"] = source

	u.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveUnit(u); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

func (s *SyzygyService) DbCheckAppend(unitID, runID string, check domain.DbCheck) (map[string]any, error) {
	u, err := s.store.GetUnit(unitID)
	if err != nil {
		return nil, err
	}
	run, err := findRun(u, runID)
	if err != nil {
		return nil, err
	}

	checkID, err := domain.NewID("db")
	if err != nil {
		return nil, err
	}
	check.CheckID = checkID
	run.DBChecks = append(run.DBChecks, &check)
	u.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveUnit(u); err != nil {
		return nil, err
	}

	return map[string]any{"dbcheck_id": checkID}, nil
}

func (s *SyzygyService) GetUnit(unitID string) (*domain.Unit, error) {
	return s.store.GetUnit(unitID)
}

func (s *SyzygyService) ListUnitIDs() ([]string, error) {
	base := s.store.BaseDir()
	if base == "" {
		base = os.Getenv("SYZYGY_DATA_DIR")
	}
	if base == "" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			base = filepath.Join(home, ".syzygy-data")
		}
	}
	if base == "" {
		return []string{}, nil
	}

	unitsDir := filepath.Join(base, "units")
	entries, err := os.ReadDir(unitsDir)
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

func (s *SyzygyService) SetUnitMeta(unitID string, meta map[string]any) (map[string]any, error) {
	u, err := s.store.GetOrCreateUnit(unitID, "", nil)
	if err != nil {
		return nil, err
	}
	if u.Meta == nil {
		u.Meta = map[string]any{}
	}
	for k, v := range meta {
		u.Meta[k] = v
	}
	u.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveUnit(u); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

func findRun(u *domain.Unit, runID string) (*domain.Run, error) {
	for _, r := range u.Runs {
		if r.RunID == runID {
			return r, nil
		}
	}
	return nil, NewAppError("run_not_found", "run not found")
}
