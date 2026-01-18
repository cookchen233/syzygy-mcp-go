package application

import "github.com/cookchen233/syzygy-mcp-go/internal/domain"

type Store interface {
	GetOrCreateUnit(projectKey string, unitID string, title string, env map[string]any) (*domain.Unit, error)
	GetUnit(projectKey string, unitID string) (*domain.Unit, error)
	SaveUnit(projectKey string, u *domain.Unit) error
	ListUnitIDs(projectKey string) ([]string, error)
	BaseDir() string
}
