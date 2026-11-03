package model

import (
	"time"
)

// ResourceSearch 记录一次由 Agent 发起的外部资源检索。
type ResourceSearch struct {
	// Timestamps 保存检索记录的生命周期字段。
	Timestamps
	// OwnerID 是检索所属的系统用户 ID。
	OwnerID uint `json:"-" gorm:"not null;index:idx_xy_resource_searches_owner"`
	// Query 是用户或 Agent 提交的检索条件。
	Query string `json:"query" gorm:"size:500;not null"`
	// Status 是检索执行状态。
	Status string `json:"status" gorm:"size:32;not null;default:running"`
	// Directives 是检索过程产生的前端指令。
	Directives string `json:"directives" gorm:"type:text"`
}

// TableName 返回资源检索表名。
func (ResourceSearch) TableName() string { return "xy_resource_searches" }

// ResourceCandidate 表示 MCP 检索返回的一个待选择资源。
type ResourceCandidate struct {
	// Timestamps 保存候选资源的生命周期字段。
	Timestamps
	// SearchID 是候选资源所属的检索 ID。
	SearchID uint `json:"searchId" gorm:"not null;index;uniqueIndex:idx_xy_candidate_external"`
	// MCPServerCode 是候选资源来源的 MCP 服务编码。
	MCPServerCode string `json:"mcpServerCode" gorm:"size:64;not null;uniqueIndex:idx_xy_candidate_external"`
	// ExternalID 是来源系统中的资源标识。
	ExternalID string `json:"externalId" gorm:"size:1000;not null;uniqueIndex:idx_xy_candidate_external"`
	// Name 是资源显示名称。
	Name string `json:"name" gorm:"size:500;not null"`
	// Size 是资源字节数。
	Size int64 `json:"size" gorm:"not null;default:0"`
	// MIMEType 是资源媒体类型。
	MIMEType string `json:"mimeType" gorm:"size:200"`
	// Source 是资源来源说明或地址。
	Source string `json:"source" gorm:"type:text;not null"`
	// Preview 是资源预览信息。
	Preview string `json:"preview" gorm:"type:text"`
	// License 是来源声明的许可类型。
	License string `json:"license" gorm:"size:100;not null;default:unknown"`
	// Risk 是资源风险提示。
	Risk string `json:"risk" gorm:"size:1000"`
	// ExpiresAt 是候选资源失效时间。
	ExpiresAt *time.Time `json:"expiresAt"`
}

// TableName 返回资源候选表名。
func (ResourceCandidate) TableName() string { return "xy_resource_candidates" }

// ResourceTask 记录资源获取、处理与上传的长任务状态。
type ResourceTask struct {
	// Timestamps 保存资源任务的生命周期字段。
	Timestamps
	// OwnerID 是任务所属的系统用户 ID。
	OwnerID uint `json:"-" gorm:"not null;index:idx_xy_resource_tasks_owner;uniqueIndex:idx_xy_resource_task_idempotency"`
	// AgentRunID 是发起任务的 Agent 运行 ID。
	AgentRunID *uint `json:"-" gorm:"index"`
	// CandidateID 是任务选用的候选资源 ID。
	CandidateID uint `json:"candidateId" gorm:"not null;index"`
	// MCPServerCode 是执行任务的 MCP 服务编码。
	MCPServerCode string `json:"mcpServerCode" gorm:"size:64;not null"`
	// ExternalTaskID 是 MCP 服务返回的任务标识。
	ExternalTaskID string `json:"externalTaskId" gorm:"size:200;index"`
	// IdempotencyKey 防止同一用户重复创建资源任务。
	IdempotencyKey string `json:"-" gorm:"size:100;not null;uniqueIndex:idx_xy_resource_task_idempotency"`
	// State 是任务整体状态。
	State string `json:"state" gorm:"size:32;not null;default:queued;index"`
	// Stage 是当前处理阶段。
	Stage string `json:"stage" gorm:"size:64"`
	// ProgressPercent 是零到一百的完成百分比。
	ProgressPercent int `json:"progressPercent" gorm:"not null;default:0"`
	// ProgressDone 是已完成工作量。
	ProgressDone int64 `json:"progressDone" gorm:"not null;default:0"`
	// ProgressTotal 是总工作量。
	ProgressTotal int64 `json:"progressTotal" gorm:"not null;default:0"`
	// ProgressUnit 是进度工作量单位。
	ProgressUnit string `json:"progressUnit" gorm:"size:32"`
	// Result 是 MCP 返回的任务结果。
	Result string `json:"result" gorm:"type:text"`
	// ErrorCode 是稳定错误编码。
	ErrorCode string `json:"errorCode" gorm:"size:64"`
	// ErrorMessage 是任务失败时的安全错误描述。
	ErrorMessage string `json:"errorMessage" gorm:"size:2000"`
	// StartedAt 是任务开始处理时间。
	StartedAt *time.Time `json:"startedAt"`
	// CompletedAt 是任务结束时间。
	CompletedAt *time.Time `json:"completedAt"`
	// DeliveryTokenHash 是一次性交付令牌摘要。
	DeliveryTokenHash string `json:"-" gorm:"size:64"`
	// DeliveryExpiresAt 是一次性交付令牌到期时间。
	DeliveryExpiresAt *time.Time `json:"-"`
	// UploadObjectKey 是任务上传对象的内部存储键。
	UploadObjectKey string `json:"-" gorm:"size:1000"`
}

// TableName 返回资源任务表名。
func (ResourceTask) TableName() string { return "xy_resource_tasks" }

// Resource 表示已进入本地对象存储的可用资源。
type Resource struct {
	// Timestamps 保存资源记录的生命周期字段。
	Timestamps
	// OwnerID 是资源所属的系统用户 ID。
	OwnerID uint `json:"-" gorm:"not null;index:idx_xy_resources_owner"`
	// TaskID 是生成资源的任务 ID。
	TaskID uint `json:"taskId" gorm:"not null;index"`
	// Name 是资源显示名称。
	Name string `json:"name" gorm:"size:500;not null"`
	// Size 是资源字节数。
	Size int64 `json:"size" gorm:"not null"`
	// MIMEType 是资源媒体类型。
	MIMEType string `json:"mimeType" gorm:"size:200"`
	// SHA256 是资源内容摘要。
	SHA256 string `json:"sha256" gorm:"size:64;index"`
	// Bucket 是内部对象存储桶名称。
	Bucket string `json:"-" gorm:"size:200;not null"`
	// ObjectKey 是内部对象存储键。
	ObjectKey string `json:"-" gorm:"size:1000;not null;uniqueIndex"`
	// State 是资源可用性状态。
	State string `json:"state" gorm:"size:32;not null;default:available"`
}

// TableName 返回本地资源表名。
func (Resource) TableName() string { return "xy_resources" }

// ResourceSource 保存本地资源与外部来源的追溯关系。
type ResourceSource struct {
	// Timestamps 保存来源记录的生命周期字段。
	Timestamps
	// ResourceID 是本地资源 ID。
	ResourceID uint `json:"resourceId" gorm:"not null;index;uniqueIndex:idx_xy_resource_source_external"`
	// MCPServerCode 是资源来源的 MCP 服务编码。
	MCPServerCode string `json:"mcpServerCode" gorm:"size:64;not null;uniqueIndex:idx_xy_resource_source_external"`
	// ExternalID 是来源系统中的资源标识。
	ExternalID string `json:"externalId" gorm:"size:1000;not null;uniqueIndex:idx_xy_resource_source_external"`
	// SourceName 是资源来源显示名称。
	SourceName string `json:"sourceName" gorm:"size:500"`
	// SourceReference 是可追溯的来源引用。
	SourceReference string `json:"sourceReference" gorm:"size:1000"`
	// Metadata 是来源附加信息的序列化值。
	Metadata string `json:"metadata" gorm:"type:text"`
}

// TableName 返回资源来源表名。
func (ResourceSource) TableName() string { return "xy_resource_sources" }
