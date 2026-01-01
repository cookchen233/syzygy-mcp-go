package application

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func (s *SyzygyService) Crystallize(unitID, runID, template, outputDir string) (map[string]any, error) {
	u, err := s.store.GetUnit(unitID)
	if err != nil {
		return nil, err
	}
	run, err := findRun(u, runID)
	if err != nil {
		return nil, err
	}

	if outputDir == "" {
		outputDir = filepath.Join("./syzygy-artifacts", unitID, runID)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, err
	}

	if template == "" {
		template = "spec_json"
	}

	paths := map[string]string{}

	// 1) Save spec
	specPath := filepath.Join(outputDir, "spec.json")
	b, err := json.MarshalIndent(map[string]any{
		"unit_id": unitID,
		"run_id":  runID,
		"steps":   run.Steps,
		"anchors": run.Anchors,
		"db_checks": run.DBChecks,
		"variables": run.Variables,
		"env":       u.Env,
	}, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(specPath, b, 0o644); err != nil {
		return nil, err
	}
	paths["spec"] = specPath

	// 2) Save a minimal Playwright TS template (optional)
	pwPath := filepath.Join(outputDir, "e2e.spec.ts")
	pwContent := []byte("import { test, expect } from '@playwright/test'\n\n" +
		"test('SYZYGY unit " + unitID + "', async ({ page }) => {\n" +
		"  // TODO: load spec.json and execute steps via your runner\n" +
		"  await page.goto(process.env.BASE_URL || 'http://localhost');\n" +
		"  await expect(page).toBeTruthy();\n" +
		"});\n")
	_ = os.WriteFile(pwPath, pwContent, 0o644)
	paths["playwright_ts"] = pwPath

	run.Artifacts = paths
	u.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveUnit(u); err != nil {
		return nil, err
	}

	return map[string]any{"artifact_paths": paths}, nil
}

func (s *SyzygyService) Replay(unitID, runID, command string, args []string, cwd string, env map[string]any) (map[string]any, error) {
	u, err := s.store.GetUnit(unitID)
	if err != nil {
		return nil, err
	}
	run, err := findRun(u, runID)
	if err != nil {
		return nil, err
	}

	if command == "" {
		specPath := ""
		if run.Artifacts != nil {
			specPath = run.Artifacts["spec"]
		}
		if specPath == "" {
			return nil, NewAppError("missing_artifact", "spec artifact not found; run syzygy_crystallize first")
		}
		// Default: call node runner (parameterized)
		command = "node"
		args = []string{"./runner-node/bin/syzygy-runner.js", specPath}
		if cwd == "" {
			cwd = "./"
		}
		if env == nil {
			env = map[string]any{}
		}
		env["SYZYGY_SPEC"] = specPath
	}

	cmd := exec.Command(command, args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	cmd.Env = os.Environ()
	for k, v := range u.Env {
		if sv, ok := v.(string); ok {
			cmd.Env = append(cmd.Env, k+"="+sv)
		}
	}
	for k, v := range env {
		if sv, ok := v.(string); ok {
			cmd.Env = append(cmd.Env, k+"="+sv)
		}
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]any{"ok": false, "output": string(out), "error": err.Error(), "anchors": run.Anchors}, nil
	}
	return map[string]any{"ok": true, "output": string(out), "anchors": run.Anchors}, nil
}
