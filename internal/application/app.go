package application

import "log"

type App struct {
	tools  *ToolRegistry
	logger *log.Logger
}

func NewApp(store Store, logger *log.Logger) *App {
	if logger == nil {
		logger = log.Default()
	}

	svc := NewSyzygyService(store, logger)
	tools := NewToolRegistry(svc)
	return &App{tools: tools, logger: logger}
}

func (a *App) ToolRegistry() *ToolRegistry {
	return a.tools
}
