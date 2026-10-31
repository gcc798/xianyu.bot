package initialize

import (
	"context"
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model"
)

func Database(ctx context.Context) error {
	if global.GVA_DB == nil {
		return errors.New("database is not initialized")
	}

	db := global.GVA_DB.WithContext(ctx)
	if global.GVA_CONFIG.System.DbType == "sqlite" {
		for _, statement := range []string{
			"PRAGMA journal_mode=WAL",
			"PRAGMA busy_timeout=5000",
			"PRAGMA foreign_keys=ON",
		} {
			if err := db.Exec(statement).Error; err != nil {
				return err
			}
		}
	}
	if err := db.AutoMigrate(&model.Account{}, &model.Credential{}, &model.Setting{}, &model.Product{}, &model.ProductSetting{}, &model.ReplyRule{}, &model.ReplyContent{}, &model.DeliveryRule{}, &model.DeliveryResource{}, &model.DeliveryTask{}, &model.RelistTask{}, &model.Message{}, &model.ConversationTakeover{}, &model.ReplyTask{}, &model.Order{}, &model.OrderEvent{}, &model.BusinessEvent{}, &model.AIModel{}, &model.MCPServer{}, &model.AgentConfig{}, &model.ResourceSearch{}, &model.ResourceCandidate{}, &model.ResourceTask{}, &model.Resource{}, &model.ResourceSource{}, &model.AgentSession{}, &model.AgentMessage{}, &model.AgentRun{}, &model.AgentRunTool{}, &model.AgentEvent{}); err != nil {
		return err
	}
	config := model.DefaultAgentConfig()
	return db.FirstOrCreate(&config, model.AgentConfig{ID: 1}).Error
}
