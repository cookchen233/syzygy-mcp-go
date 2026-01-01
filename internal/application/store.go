package application

import "github.com/cookchen233/syzygy-mcp-go/internal/domain"

type Store interface {
	GetOrCreateUnit(unitID string, title string, env map[string]any) (*domain.Unit, error)
	GetUnit(unitID string) (*domain.Unit, error)
	SaveUnit(u *domain.Unit) error
	BaseDir() string
}
