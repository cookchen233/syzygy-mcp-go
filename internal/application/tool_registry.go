package application

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cookchen233/syzygy-mcp-go/internal/domain"
)

type ToolRegistry struct {
	svc *SyzygyService
}

func NewToolRegistry(svc *SyzygyService) *ToolRegistry {
	return &ToolRegistry{svc: svc}
}

func (r *ToolRegistry) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "syzygy_project_init",
			Description: "Initialize project runtime config (初始化项目运行配置：DB/BASE_URL/artifacts/runner)",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"project_key":    map[string]any{"type": "string"},
					"env":            map[string]any{"type": "object"},
					"runner_command": map[string]any{"type": "string"},
					"runner_dir":     map[string]any{"type": "string"},
				},
				"required": []string{},
			},
		},
		{
			Name:        "syzygy_unit_start",
			Description: "Start a Syzygy unit run (创建并开始一个单元 run)",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"unit_id":   map[string]any{"type": "string"},
					"title":     map[string]any{"type": "string"},
					"env":       map[string]any{"type": "object"},
					"variables": map[string]any{"type": "object"},
				},
				"required": []string{"unit_id"},
			},
		},
		{
			Name:        "syzygy_unit_meta_set",
			Description: "Set unit meta (设置单元元数据/触点)",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"unit_id": map[string]any{"type": "string"},
					"meta":    map[string]any{"type": "object"},
				},
				"required": []string{"unit_id", "meta"},
			},
		},
		{
			Name:        "syzygy_unit_meta_set_json",
			Description: "Set unit meta by JSON string (设置单元元数据 - JSON 字符串)",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"unit_id":     map[string]any{"type": "string"},
					"meta":        map[string]any{"type": "object"},
					"meta_json":   map[string]any{"type": "string"},
					"meta_base64": map[string]any{"type": "string"},
				},
				"required": []string{"unit_id"},
			},
		},
		{
			Name:        "syzygy_plan_impacted_units",
			Description: "Plan impacted units by changed files/APIs/tables (根据改动规划需要回放的单元)",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"changed_files":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"changed_apis":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"changed_tables": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"tags":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
				"required": []string{},
			},
		},
		{
			Name:        "syzygy_step_append",
			Description: "Append an action step (追加动作步骤)",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"unit_id": map[string]any{"type": "string"},
					"run_id":  map[string]any{"type": "string"},
					"step":    map[string]any{"type": "object"},
				},
				"required": []string{"unit_id", "run_id", "step"},
			},
		},
		{
			Name:        "syzygy_step_append_json",
			Description: "Append an action step by JSON string (追加动作步骤 - JSON 字符串)",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"unit_id":     map[string]any{"type": "string"},
					"run_id":      map[string]any{"type": "string"},
					"step_json":   map[string]any{"type": "string"},
					"step":        map[string]any{"type": "object"},
					"step_base64": map[string]any{"type": "string"},
				},
				"required": []string{"unit_id", "run_id"},
			},
		},
		{
			Name:        "syzygy_steps_append_batch",
			Description: "Append action steps in batch (批量追加动作步骤)",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"unit_id": map[string]any{"type": "string"},
					"run_id":  map[string]any{"type": "string"},
					"steps": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "object"},
					},
				},
				"required": []string{"unit_id", "run_id", "steps"},
			},
		},
		{
			Name:        "syzygy_anchor_set",
			Description: "Set an anchor value (设置数据锚点)",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"unit_id": map[string]any{"type": "string"},
					"run_id":  map[string]any{"type": "string"},
					"key":     map[string]any{"type": "string"},
					"value":   map[string]any{"type": "string"},
					"source":  map[string]any{"type": "string"},
				},
				"required": []string{"unit_id", "run_id", "key", "value"},
			},
		},
		{
			Name:        "syzygy_dbcheck_append",
			Description: "Append a DB check (追加数据库断言)",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"unit_id":  map[string]any{"type": "string"},
					"run_id":   map[string]any{"type": "string"},
					"db_check": map[string]any{"type": "object"},
				},
				"required": []string{"unit_id", "run_id", "db_check"},
			},
		},
		{
			Name:        "syzygy_crystallize",
			Description: "Generate artifacts (生成固化产物)",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"unit_id":    map[string]any{"type": "string"},
					"run_id":     map[string]any{"type": "string"},
					"template":   map[string]any{"type": "string"},
					"output_dir": map[string]any{"type": "string"},
				},
				"required": []string{"unit_id", "run_id"},
			},
		},
		{
			Name:        "syzygy_replay",
			Description: "Replay a crystallized unit (回放固化用例)",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"unit_id": map[string]any{"type": "string"},
					"run_id":  map[string]any{"type": "string"},
					"command": map[string]any{"type": "string"},
					"args": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
					"cwd": map[string]any{"type": "string"},
					"env": map[string]any{"type": "object"},
				},
				"required": []string{"unit_id", "run_id"},
			},
		},
		{
			Name:        "syzygy_selfcheck",
			Description: "Self-check a unit run for SYZYGY compliance (自查单元运行是否符合SYZYGY规范)。完成开发后必须调用此工具验证：1.固化是否完成 2.回放是否执行且成功 3.三层对齐是否达成 4.交付格式是否正确",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"unit_id": map[string]any{"type": "string"},
					"run_id":  map[string]any{"type": "string"},
				},
				"required": []string{"unit_id", "run_id"},
			},
		},
	}
}

func latestRunID(u *domain.Unit) string {
	if u == nil || len(u.Runs) == 0 {
		return ""
	}
	return u.Runs[len(u.Runs)-1].RunID
}

func toStringSliceAny(v any) []string {
	out := []string{}
	arr, ok := v.([]any)
	if !ok {
		return out
	}
	for _, it := range arr {
		if s, ok := it.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func matchesAny(hay string, needles []string) bool {
	for _, n := range needles {
		if n == "" {
			continue
		}
		if strings.Contains(hay, n) {
			return true
		}
	}
	return false
}

func parseActionStepFromMap(stepRaw map[string]any) domain.ActionStep {
	step := domain.ActionStep{}
	if v, ok := stepRaw["name"].(string); ok {
		step.Name = v
	}
	if v, ok := stepRaw["util"].(map[string]any); ok {
		step.Util = v
	}
	if v, ok := stepRaw["db"].(map[string]any); ok {
		step.DB = v
	}
	if v, ok := stepRaw["ui"].(map[string]any); ok {
		step.UI = v
	}
	if v, ok := stepRaw["net"].(map[string]any); ok {
		step.Net = v
	}
	if v, ok := stepRaw["expect"].(map[string]any); ok {
		step.Expect = v
	}
	return step
}

func (r *ToolRegistry) CallTool(name string, args map[string]any) (any, error) {
	switch name {
	case "syzygy_project_init":
		projectKey, _ := args["project_key"].(string)
		env, _ := args["env"].(map[string]any)
		runnerCommand, _ := args["runner_command"].(string)
		runnerDir, _ := args["runner_dir"].(string)
		if env == nil {
			env = map[string]any{}
		}
		return r.svc.ProjectInit(projectKey, env, runnerCommand, runnerDir)
	case "syzygy_unit_start":
		unitID, _ := args["unit_id"].(string)
		title, _ := args["title"].(string)
		env, _ := args["env"].(map[string]any)
		vars, _ := args["variables"].(map[string]any)
		if unitID == "" {
			return nil, NewAppError("invalid_unit_id", "unit_id is required")
		}
		return r.svc.UnitStart(unitID, title, env, vars)
	case "syzygy_unit_meta_set":
		unitID, _ := args["unit_id"].(string)
		meta, ok := args["meta"].(map[string]any)
		if unitID == "" || !ok {
			return nil, NewAppError("invalid_args", "unit_id and meta are required")
		}
		return r.svc.SetUnitMeta(unitID, meta)
	case "syzygy_unit_meta_set_json":
		unitID, _ := args["unit_id"].(string)
		if unitID == "" {
			return nil, NewAppError("invalid_args", "unit_id is required")
		}
		if meta, ok := args["meta"].(map[string]any); ok {
			return r.svc.SetUnitMeta(unitID, meta)
		}
		metaJSON, _ := args["meta_json"].(string)
		if metaJSON == "" {
			if b64, _ := args["meta_base64"].(string); b64 != "" {
				decoded, err := base64.StdEncoding.DecodeString(b64)
				if err != nil {
					return nil, NewAppError("invalid_meta_base64", fmt.Sprintf("invalid meta_base64: %v", err))
				}
				metaJSON = string(decoded)
			}
		}
		if metaJSON == "" {
			return nil, NewAppError("invalid_args", "missing meta. Provide meta (object) or meta_json (string) or meta_base64 (string)")
		}
		var meta map[string]any
		if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
			return nil, NewAppError("invalid_meta_json", fmt.Sprintf("invalid meta_json: %v", err))
		}
		return r.svc.SetUnitMeta(unitID, meta)
	case "syzygy_plan_impacted_units":
		changedFiles := toStringSliceAny(args["changed_files"])
		changedApis := toStringSliceAny(args["changed_apis"])
		changedTables := toStringSliceAny(args["changed_tables"])
		wantedTags := toStringSliceAny(args["tags"])

		unitIDs, err := r.svc.ListUnitIDs()
		if err != nil {
			return nil, err
		}

		out := []map[string]any{}
		for _, uid := range unitIDs {
			u, err := r.svc.GetUnit(uid)
			if err != nil {
				continue
			}
			touch, _ := u.Meta["touchpoints"].(map[string]any)
			apiArr := toStringSliceAny(touch["api"])
			tableArr := toStringSliceAny(touch["db_tables"])
			fileArr := toStringSliceAny(touch["files"])
			tagArr := toStringSliceAny(u.Meta["tags"])

			reasons := []string{}
			for _, f := range changedFiles {
				if matchesAny(f, fileArr) {
					reasons = append(reasons, "file:"+f)
					break
				}
			}
			for _, a := range changedApis {
				if matchesAny(a, apiArr) {
					reasons = append(reasons, "api:"+a)
					break
				}
			}
			for _, t := range changedTables {
				if matchesAny(t, tableArr) {
					reasons = append(reasons, "table:"+t)
					break
				}
			}
			if len(wantedTags) > 0 {
				if matchesAny(strings.Join(tagArr, ","), wantedTags) {
					reasons = append(reasons, "tag")
				}
			}

			if len(reasons) > 0 {
				out = append(out, map[string]any{
					"unit_id": uid,
					"title":   u.Title,
					"reasons": reasons,
				})
			}
		}
		return map[string]any{"impacted_units": out}, nil
	case "syzygy_step_append":
		unitID, _ := args["unit_id"].(string)
		runID, _ := args["run_id"].(string)
		if runID == "" {
			u, err := r.svc.GetUnit(unitID)
			if err == nil {
				runID = latestRunID(u)
			}
		}
		stepRaw, ok := args["step"].(map[string]any)
		if !ok {
			return nil, NewAppError("invalid_step", "step must be object; missing or wrong type")
		}
		step := parseActionStepFromMap(stepRaw)
		return r.svc.StepAppend(unitID, runID, step)
	case "syzygy_step_append_json":
		unitID, _ := args["unit_id"].(string)
		runID, _ := args["run_id"].(string)
		if runID == "" {
			u, err := r.svc.GetUnit(unitID)
			if err == nil {
				runID = latestRunID(u)
			}
		}
		// Prefer step object if provided
		if stepRaw, ok := args["step"].(map[string]any); ok {
			step := parseActionStepFromMap(stepRaw)
			return r.svc.StepAppend(unitID, runID, step)
		}

		stepJSON, _ := args["step_json"].(string)
		if stepJSON == "" {
			if b64, _ := args["step_base64"].(string); b64 != "" {
				decoded, err := base64.StdEncoding.DecodeString(b64)
				if err != nil {
					return nil, NewAppError("invalid_step_base64", fmt.Sprintf("invalid step_base64: %v", err))
				}
				stepJSON = string(decoded)
			}
		}
		if stepJSON == "" {
			return nil, NewAppError("invalid_step", "missing step. Provide step (object) or step_json (string) or step_base64 (string)")
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(stepJSON), &raw); err != nil {
			return nil, NewAppError("invalid_step_json", fmt.Sprintf("invalid step_json: %v", err))
		}
		step := parseActionStepFromMap(raw)
		return r.svc.StepAppend(unitID, runID, step)
	case "syzygy_steps_append_batch":
		unitID, _ := args["unit_id"].(string)
		runID, _ := args["run_id"].(string)
		if runID == "" {
			u, err := r.svc.GetUnit(unitID)
			if err == nil {
				runID = latestRunID(u)
			}
		}
		arr, ok := args["steps"].([]any)
		if !ok {
			return nil, NewAppError("invalid_steps", "steps must be array")
		}
		stepIDs := []string{}
		for _, it := range arr {
			m, ok := it.(map[string]any)
			if !ok {
				return nil, NewAppError("invalid_steps", "each step must be object")
			}
			step := parseActionStepFromMap(m)
			res, err := r.svc.StepAppend(unitID, runID, step)
			if err != nil {
				return nil, err
			}
			if id, ok := res["step_id"].(string); ok {
				stepIDs = append(stepIDs, id)
			}
		}
		return map[string]any{"step_ids": stepIDs}, nil
	case "syzygy_anchor_set":
		unitID, _ := args["unit_id"].(string)
		runID, _ := args["run_id"].(string)
		key, _ := args["key"].(string)
		value, _ := args["value"].(string)
		source, _ := args["source"].(string)
		return r.svc.AnchorSet(unitID, runID, key, value, source)
	case "syzygy_dbcheck_append":
		unitID, _ := args["unit_id"].(string)
		runID, _ := args["run_id"].(string)
		if runID == "" {
			u, err := r.svc.GetUnit(unitID)
			if err == nil {
				runID = latestRunID(u)
			}
		}
		checkRaw, ok := args["db_check"].(map[string]any)
		if !ok {
			return nil, NewAppError("invalid_db_check", "db_check must be object")
		}
		check := domain.DbCheck{Params: map[string]string{}, Assert: map[string]any{}}
		if v, ok := checkRaw["name"].(string); ok {
			check.Name = v
		}
		if v, ok := checkRaw["dms"].(string); ok {
			check.DMS = v
		}
		if v, ok := checkRaw["sql"].(string); ok {
			check.SQL = v
		}
		if v, ok := checkRaw["params"].(map[string]any); ok {
			for k, vv := range v {
				if s, ok := vv.(string); ok {
					check.Params[k] = s
				}
			}
		}
		if v, ok := checkRaw["assert"].(map[string]any); ok {
			check.Assert = v
		}
		return r.svc.DbCheckAppend(unitID, runID, check)
	case "syzygy_crystallize":
		unitID, _ := args["unit_id"].(string)
		runID, _ := args["run_id"].(string)
		if runID == "" {
			u, err := r.svc.GetUnit(unitID)
			if err == nil {
				runID = latestRunID(u)
			}
		}
		tpl, _ := args["template"].(string)
		outDir, _ := args["output_dir"].(string)
		if unitID == "" || runID == "" {
			return nil, NewAppError("invalid_args", "unit_id and run_id are required")
		}
		return r.svc.Crystallize(unitID, runID, tpl, outDir)
	case "syzygy_replay":
		unitID, _ := args["unit_id"].(string)
		runID, _ := args["run_id"].(string)
		if runID == "" {
			u, err := r.svc.GetUnit(unitID)
			if err == nil {
				runID = latestRunID(u)
			}
		}
		cmd, _ := args["command"].(string)
		rawArgs, _ := args["args"].([]any)
		argv := []string{}
		for _, v := range rawArgs {
			if s, ok := v.(string); ok {
				argv = append(argv, s)
			}
		}
		cwd, _ := args["cwd"].(string)
		env, _ := args["env"].(map[string]any)
		if unitID == "" || runID == "" {
			return nil, NewAppError("invalid_args", "unit_id and run_id are required")
		}
		return r.svc.Replay(unitID, runID, cmd, argv, cwd, env)
	case "syzygy_selfcheck":
		unitID, _ := args["unit_id"].(string)
		runID, _ := args["run_id"].(string)
		if runID == "" {
			u, err := r.svc.GetUnit(unitID)
			if err == nil {
				runID = latestRunID(u)
			}
		}
		if unitID == "" || runID == "" {
			return nil, NewAppError("invalid_args", "unit_id and run_id are required")
		}
		return r.svc.SelfCheck(unitID, runID)
	default:
		return nil, NewAppError("tool_not_implemented", "tool not implemented: "+name)
	}
}
