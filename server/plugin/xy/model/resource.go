package model

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type ResourceSearch struct {
	global.GVA_MODEL
	OwnerID    uint   `json:"-" gorm:"not null;index:idx_xy_resource_searches_owner"`
	Query      string `json:"query" gorm:"size:500;not null"`
	Status     string `json:"status" gorm:"size:32;not null;default:running"`
	Directives string `json:"directives" gorm:"type:text"`
}

func (ResourceSearch) TableName() string { return "xy_resource_searches" }

type ResourceCandidate struct {
	global.GVA_MODEL
	SearchID      uint       `json:"searchId" gorm:"not null;index;uniqueIndex:idx_xy_candidate_external"`
	MCPServerCode string     `json:"mcpServerCode" gorm:"size:64;not null;uniqueIndex:idx_xy_candidate_external"`
	ExternalID    string     `json:"externalId" gorm:"size:1000;not null;uniqueIndex:idx_xy_candidate_external"`
	Name          string     `json:"name" gorm:"size:500;not null"`
	Size          int64      `json:"size" gorm:"not null;default:0"`
	MIMEType      string     `json:"mimeType" gorm:"size:200"`
	Source        string     `json:"source" gorm:"type:text;not null"`
	Preview       string     `json:"preview" gorm:"type:text"`
	License       string     `json:"license" gorm:"size:100;not null;default:unknown"`
	Risk          string     `json:"risk" gorm:"size:1000"`
	ExpiresAt     *time.Time `json:"expiresAt"`
}

func (ResourceCandidate) TableName() string { return "xy_resource_candidates" }

type ResourceTask struct {
	global.GVA_MODEL
	OwnerID         uint       `json:"-" gorm:"not null;index:idx_xy_resource_tasks_owner;uniqueIndex:idx_xy_resource_task_idempotency"`
	CandidateID     uint       `json:"candidateId" gorm:"not null;index"`
	MCPServerCode   string     `json:"mcpServerCode" gorm:"size:64;not null"`
	ExternalTaskID  string     `json:"externalTaskId" gorm:"size:200;index"`
	IdempotencyKey  string     `json:"-" gorm:"size:100;not null;uniqueIndex:idx_xy_resource_task_idempotency"`
	State           string     `json:"state" gorm:"size:32;not null;default:queued;index"`
	Stage           string     `json:"stage" gorm:"size:64"`
	ProgressPercent int        `json:"progressPercent" gorm:"not null;default:0"`
	ProgressDone    int64      `json:"progressDone" gorm:"not null;default:0"`
	ProgressTotal   int64      `json:"progressTotal" gorm:"not null;default:0"`
	ProgressUnit    string     `json:"progressUnit" gorm:"size:32"`
	Result          string     `json:"result" gorm:"type:text"`
	ErrorCode       string     `json:"errorCode" gorm:"size:64"`
	ErrorMessage    string     `json:"errorMessage" gorm:"size:2000"`
	StartedAt       *time.Time `json:"startedAt"`
	CompletedAt     *time.Time `json:"completedAt"`
}

func (ResourceTask) TableName() string { return "xy_resource_tasks" }

type Resource struct {
	global.GVA_MODEL
	OwnerID   uint   `json:"-" gorm:"not null;index:idx_xy_resources_owner"`
	TaskID    uint   `json:"taskId" gorm:"not null;index"`
	Name      string `json:"name" gorm:"size:500;not null"`
	Size      int64  `json:"size" gorm:"not null"`
	MIMEType  string `json:"mimeType" gorm:"size:200"`
	SHA256    string `json:"sha256" gorm:"size:64;index"`
	Bucket    string `json:"-" gorm:"size:200;not null"`
	ObjectKey string `json:"-" gorm:"size:1000;not null;uniqueIndex"`
	State     string `json:"state" gorm:"size:32;not null;default:available"`
}

func (Resource) TableName() string { return "xy_resources" }

type ResourceSource struct {
	global.GVA_MODEL
	ResourceID      uint   `json:"resourceId" gorm:"not null;index;uniqueIndex:idx_xy_resource_source_external"`
	MCPServerCode   string `json:"mcpServerCode" gorm:"size:64;not null;uniqueIndex:idx_xy_resource_source_external"`
	ExternalID      string `json:"externalId" gorm:"size:1000;not null;uniqueIndex:idx_xy_resource_source_external"`
	SourceName      string `json:"sourceName" gorm:"size:500"`
	SourceReference string `json:"sourceReference" gorm:"size:1000"`
	Metadata        string `json:"metadata" gorm:"type:text"`
}

func (ResourceSource) TableName() string { return "xy_resource_sources" }
