package agent

import (
	"context"
	"errors"

	"github.com/gcc798/xianyu.bot/internal/model"
	"gorm.io/gorm"
)

func Initialize(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return errors.New("database is not initialized")
	}
	config := model.DefaultAgentConfig()
	return db.WithContext(ctx).FirstOrCreate(&config, model.AgentConfig{ID: 1}).Error
}

func NewAIConfig(db *gorm.DB) *AIConfig {
	return &AIConfig{db: db}
}

func NewSession(db *gorm.DB) *Session {
	return &Session{db: db}
}

func NewMCPRegistry(db *gorm.DB) *MCPRegistry {
	return &MCPRegistry{db: db}
}

func NewResourceCatalog(db *gorm.DB, events *EventHub, registry *MCPRegistry) *ResourceCatalog {
	return &ResourceCatalog{db: db, events: events, registry: registry}
}

func NewRunner(ctx context.Context, db *gorm.DB, registry *MCPRegistry, resources *ResourceCatalog, events *EventHub) *Runner {
	return &Runner{
		root: rootContext(ctx), db: db, registry: registry, resources: resources, events: events,
		workers: make(map[uint]context.CancelFunc),
	}
}

func rootContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}
