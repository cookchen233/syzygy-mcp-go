package domain

import (
	"time"
)

type Unit struct {
	UnitID    string         `json:"unit_id"`
	Title     string         `json:"title"`
	Env       map[string]any `json:"env"`
	Meta      map[string]any `json:"meta,omitempty"`
	Runs      []*Run         `json:"runs"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type Run struct {
	RunID     string                 `json:"run_id"`
	Status    string                 `json:"status"`
	Variables map[string]any         `json:"variables"`
	Steps     []*ActionStep          `json:"steps"`
	Anchors   map[string]string      `json:"anchors"`
	DBChecks  []*DbCheck             `json:"db_checks"`
	Artifacts map[string]string      `json:"artifacts"`
	StartedAt time.Time              `json:"started_at"`
	EndedAt   *time.Time             `json:"ended_at,omitempty"`
	Meta      map[string]any         `json:"meta,omitempty"`
}

type ActionStep struct {
	StepID string         `json:"step_id"`
	Name   string         `json:"name"`
	Util   map[string]any `json:"util,omitempty"`
	DB     map[string]any `json:"db,omitempty"`
	UI     map[string]any `json:"ui,omitempty"`
	Net    map[string]any `json:"net,omitempty"`
	Expect map[string]any `json:"expect,omitempty"`
}

type DbCheck struct {
	CheckID string            `json:"check_id"`
	Name    string            `json:"name"`
	DMS     string            `json:"dms"`
	SQL     string            `json:"sql"`
	Params  map[string]string `json:"params"`
	Assert  map[string]any    `json:"assert"`
}
