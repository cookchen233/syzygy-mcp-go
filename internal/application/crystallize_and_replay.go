package application

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		"unit_id":   unitID,
		"run_id":    runID,
		"steps":     run.Steps,
		"anchors":   run.Anchors,
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

	// 环境检查和命令验证
	if err := s.validateCommand(command); err != nil {
		// 严格模式：环境问题必须明确报告，不允许mock通过
		return nil, NewAppError("environment_error", fmt.Sprintf("Command validation failed: %v. This is an environment issue that must be resolved before replay can proceed. Please check your PATH and command availability.", err))
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

	// 保存replay结果到run.Meta供selfcheck检测
	if run.Meta == nil {
		run.Meta = map[string]any{}
	}

	var result map[string]any
	if err != nil {
		result = map[string]any{"ok": false, "output": string(out), "error": err.Error(), "anchors": run.Anchors}
	} else {
		result = map[string]any{"ok": true, "output": string(out), "anchors": run.Anchors}
	}

	// 将replay结果保存到meta中
	run.Meta["replay_result"] = result
	run.Meta["replay_executed_at"] = time.Now().UTC().Format(time.RFC3339)
	u.UpdatedAt = time.Now().UTC()

	if saveErr := s.store.SaveUnit(u); saveErr != nil {
		s.logger.Printf("Warning: failed to save replay result to meta: %v", saveErr)
	}

	return result, nil
}

// validateCommand 检查命令是否存在且可执行
func (s *SyzygyService) validateCommand(command string) error {
	// 检查是否是绝对路径
	if strings.Contains(command, "/") {
		if _, err := exec.LookPath(command); err != nil {
			return fmt.Errorf("command not found at absolute path: %s", command)
		}
		return nil
	}

	// 对于相对路径命令，检查 PATH 中是否存在
	if _, err := exec.LookPath(command); err != nil {
		// 尝试常见的路径
		commonPaths := []string{"/bin", "/usr/bin", "/usr/local/bin", "/opt/homebrew/bin"}
		for _, path := range commonPaths {
			fullPath := filepath.Join(path, command)
			if _, err := os.Stat(fullPath); err == nil {
				// 找到了，但需要更新 PATH
				s.logger.Printf("Found command %s at %s, but PATH may be incomplete", command, fullPath)
				return nil
			}
		}
		return fmt.Errorf("command '%s' not found in PATH or common locations", command)
	}

	return nil
}
