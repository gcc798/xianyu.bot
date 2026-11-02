package model

import (
	"time"

	"github.com/gcc798/xianyu.bot/internal/model/common"
)

type AIModel struct {
	common.Timestamps
	Name                  string `json:"name" gorm:"size:80;not null"`
	Provider              string `json:"provider" gorm:"size:40;not null"`
	BaseURL               string `json:"baseUrl" gorm:"size:500;not null"`
	APIKey                string `json:"-" gorm:"type:text"`
	Model                 string `json:"model" gorm:"size:120;not null"`
	Enabled               bool   `json:"enabled" gorm:"not null;default:true"`
	IsDefault             bool   `json:"isDefault" gorm:"not null;default:false;index"`
	SupportsTools         bool   `json:"supportsTools" gorm:"not null;default:true"`
	SupportsStreaming     bool   `json:"supportsStreaming" gorm:"not null;default:true"`
	ContextWindow         int    `json:"contextWindow" gorm:"not null;default:128000"`
	MaxOutputTokens       int    `json:"maxOutputTokens" gorm:"not null;default:2000"`
	RequestTimeoutSeconds int    `json:"requestTimeoutSeconds" gorm:"not null;default:60"`
	ExtraConfig           string `json:"extraConfig" gorm:"type:text"`
	HasAPIKey             bool   `json:"hasApiKey" gorm:"-"`
}

func (AIModel) TableName() string { return "xy_ai_models" }

type MCPServer struct {
	common.Timestamps
	Name                  string     `json:"name" gorm:"size:80;not null"`
	Code                  string     `json:"code" gorm:"size:64;not null;uniqueIndex"`
	BaseURL               string     `json:"baseUrl" gorm:"size:500;not null"`
	AuthHeader            string     `json:"authHeader" gorm:"size:100"`
	AuthToken             string     `json:"-" gorm:"type:text"`
	Enabled               bool       `json:"enabled" gorm:"not null;default:true"`
	RequestTimeoutSeconds int        `json:"requestTimeoutSeconds" gorm:"not null;default:15"`
	TaskTimeoutSeconds    int        `json:"taskTimeoutSeconds" gorm:"not null;default:300"`
	LastStatus            string     `json:"lastStatus" gorm:"size:32;not null;default:unknown"`
	LastCheckedAt         *time.Time `json:"lastCheckedAt"`
	LastError             string     `json:"lastError" gorm:"size:1000"`
	HasAuthToken          bool       `json:"hasAuthToken" gorm:"-"`
}

func (MCPServer) TableName() string { return "xy_mcp_servers" }

type AgentConfig struct {
	ID                       uint `json:"ID" gorm:"primaryKey"`
	MaxWallTimeSeconds       int  `json:"maxWallTimeSeconds" gorm:"not null"`
	MaxIterations            int  `json:"maxIterations" gorm:"not null"`
	MaxModelCalls            int  `json:"maxModelCalls" gorm:"not null"`
	MaxToolCalls             int  `json:"maxToolCalls" gorm:"not null"`
	MaxConcurrentToolCalls   int  `json:"maxConcurrentToolCalls" gorm:"not null;default:3"`
	MaxSearchCalls           int  `json:"maxSearchCalls" gorm:"not null"`
	MaxWriteToolCalls        int  `json:"maxWriteToolCalls" gorm:"not null"`
	MaxOutputTokens          int  `json:"maxOutputTokens" gorm:"not null"`
	ConversationHistoryLimit int  `json:"conversationHistoryLimit" gorm:"not null"`
	SummaryTriggerTokens     int  `json:"summaryTriggerTokens" gorm:"not null"`
}

func (AgentConfig) TableName() string { return "xy_agent_config" }

func DefaultAgentConfig() AgentConfig {
	return AgentConfig{
		ID: 1, MaxWallTimeSeconds: 300, MaxIterations: 50, MaxModelCalls: 60,
		MaxToolCalls: 200, MaxConcurrentToolCalls: 3, MaxSearchCalls: 200, MaxWriteToolCalls: 10,
		MaxOutputTokens: 2000, ConversationHistoryLimit: 40, SummaryTriggerTokens: 24000,
	}
}
