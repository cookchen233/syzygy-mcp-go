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
	if _, err := s.EnsureProjectInitialized(); err != nil {
		return nil, err
	}

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

func (s *SyzygyService) ProjectInit(projectKey string, env map[string]any, runnerCommand string, runnerDir string) (map[string]any, error) {
	cfg := &ProjectConfig{
		ProjectKey:    projectKey,
		Env:           map[string]string{},
		RunnerCommand: normalizeRunnerCommand(runnerCommand),
		RunnerDir:     strings.TrimSpace(runnerDir),
	}
	for k, v := range env {
		cfg.Env[k] = anyToString(v)
	}
	if cfg.RunnerCommand == "" {
		cfg.RunnerCommand = "syzygy-runner"
	}
	path, err := s.SaveProjectConfig(cfg)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"ok":          true,
		"config_path": path,
		"config":      cfg,
	}, nil
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

// SelfCheck performs a comprehensive check on a unit run to verify SYZYGY compliance
func (s *SyzygyService) SelfCheck(unitID, runID string) (map[string]any, error) {
	u, err := s.store.GetUnit(unitID)
	if err != nil {
		return nil, err
	}

	run, err := findRun(u, runID)
	if err != nil {
		return nil, err
	}

	checks := []map[string]any{}
	allPassed := true

	// 4.2 开发中必查（MCP调用）- 检查run状态
	runStatusCheck := map[string]any{
		"name":     "run_status",
		"category": "development",
		"passed":   run.Status != "",
		"message":  "Run status is set",
	}
	if run.Status == "" {
		runStatusCheck["passed"] = false
		runStatusCheck["message"] = "Run status is empty"
		allPassed = false
	}
	checks = append(checks, runStatusCheck)

	// 4.3.1 固化完成检查 - 检查是否有artifacts
	crystallizeCheck := map[string]any{
		"name":     "crystallize_completed",
		"category": "completion",
		"passed":   len(run.Artifacts) > 0,
		"message":  "Crystallize has been executed",
	}
	if len(run.Artifacts) == 0 {
		crystallizeCheck["passed"] = false
		crystallizeCheck["message"] = "❌ syzygy_crystallize not executed - no artifacts found"
		allPassed = false
	}
	checks = append(checks, crystallizeCheck)

	// 4.3.2 回放验证检查 - 检查meta中是否有replay结果
	replayResult, hasReplay := run.Meta["replay_result"]
	replayCheck := map[string]any{
		"name":     "replay_verified",
		"category": "completion",
		"passed":   hasReplay,
		"message":  "Replay verification has been executed",
	}
	if !hasReplay {
		replayCheck["passed"] = false
		replayCheck["message"] = "❌ syzygy_replay not executed - no replay result found"
		allPassed = false
	} else {
		// 严格检查replay结果是否成功
		if resultMap, ok := replayResult.(map[string]any); ok {
			if okVal, exists := resultMap["ok"]; exists {
				if okBool, isBool := okVal.(bool); isBool && !okBool {
					replayCheck["passed"] = false
					errMsg := ""
					if e, hasErr := resultMap["error"].(string); hasErr {
						errMsg = e
					}
					replayCheck["message"] = "❌ syzygy_replay returned ok=false: " + errMsg
					allPassed = false
				} else {
					replayCheck["passed"] = true
					replayCheck["message"] = "✅ Replay verification successful"
				}
			}
		}
	}
	checks = append(checks, replayCheck)

	// 4.3.3 三层对齐检查 - 检查是否有UI、Net、DB相关步骤
	hasUI := false
	hasNet := false
	hasDB := false
	for _, step := range run.Steps {
		if step.UI != nil && len(step.UI) > 0 {
			hasUI = true
		}
		if step.Net != nil && len(step.Net) > 0 {
			hasNet = true
		}
		if step.DB != nil && len(step.DB) > 0 {
			hasDB = true
		}
	}
	// DB层也可以通过DBChecks来验证
	if len(run.DBChecks) > 0 {
		hasDB = true
	}

	alignmentCheck := map[string]any{
		"name":     "three_layer_alignment",
		"category": "completion",
		"passed":   hasUI || hasNet || hasDB,
		"details": map[string]bool{
			"ui":  hasUI,
			"net": hasNet,
			"db":  hasDB,
		},
		"message": "Three-layer alignment check",
	}
	if !hasUI && !hasNet && !hasDB {
		alignmentCheck["passed"] = false
		alignmentCheck["message"] = "❌ No UI/Net/DB steps found - three-layer alignment not achieved"
		allPassed = false
	}
	checks = append(checks, alignmentCheck)

	// 4.3.4 性能记录检查 - 检查是否有超时相关记录
	hasPerformanceRecord := false
	if perfData, exists := run.Meta["performance"]; exists && perfData != nil {
		hasPerformanceRecord = true
	}
	// 也检查步骤中是否有性能相关信息
	for _, step := range run.Steps {
		if step.Name != "" && (strings.Contains(step.Name, "timeout") || strings.Contains(step.Name, "performance") || strings.Contains(step.Name, "wait")) {
			hasPerformanceRecord = true
			break
		}
	}
	performanceCheck := map[string]any{
		"name":     "performance_recorded",
		"category": "completion",
		"passed":   true, // 性能记录是可选的，但如果有超时应该记录
		"message":  "Performance record check (optional)",
	}
	if !hasPerformanceRecord {
		performanceCheck["message"] = "⚠️ No performance record found (optional but recommended)"
	}
	checks = append(checks, performanceCheck)

	// 4.3.5 交付格式检查 - 检查是否有必要的元数据
	hasTitle := u.Title != ""
	formatCheck := map[string]any{
		"name":     "delivery_format",
		"category": "completion",
		"passed":   hasTitle,
		"message":  "Delivery format check",
	}
	if !hasTitle {
		formatCheck["passed"] = false
		formatCheck["message"] = "❌ Unit title is empty - delivery format incomplete"
		allPassed = false
	}
	checks = append(checks, formatCheck)

	// 汇总结果
	result := map[string]any{
		"unit_id":    unitID,
		"run_id":     runID,
		"all_passed": allPassed,
		"checks":     checks,
		"summary":    "",
	}

	if allPassed {
		result["summary"] = "🟢 SYZYGY SELFCHECK PASSED - All checks completed successfully"
	} else {
		failedChecks := []string{}
		for _, c := range checks {
			if passed, ok := c["passed"].(bool); ok && !passed {
				failedChecks = append(failedChecks, c["name"].(string))
			}
		}
		result["summary"] = "🔴 SYZYGY SELFCHECK FAILED - Failed checks: " + strings.Join(failedChecks, ", ")
	}

	return result, nil
}
