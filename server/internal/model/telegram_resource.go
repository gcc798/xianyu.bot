package model

import "time"

// TelegramDialog 保存 Go 管理的 Telegram 来源目录。
type TelegramDialog struct {
	ID                       uint       `json:"id" gorm:"primaryKey"`                                     // ID 是来源记录 ID。
	TelegramID               int64      `json:"telegramId" gorm:"uniqueIndex;not null"`                   // TelegramID 是 Telegram 会话 ID。
	AccessHash               *int64     `json:"-"`                                                        // AccessHash 是 Telegram 私有会话引用。
	Name                     string     `json:"name" gorm:"not null"`                                     // Name 是会话名称。
	Username                 string     `json:"username"`                                                 // Username 是公开用户名。
	DialogType               string     `json:"dialogType" gorm:"column:dialog_type"`                     // DialogType 是群组或频道类型。
	MemberCount              *int       `json:"memberCount"`                                              // MemberCount 是会话成员数。
	IsChannel                bool       `json:"isChannel"`                                                // IsChannel 表示频道。
	IsGroup                  bool       `json:"isGroup"`                                                  // IsGroup 表示群组。
	IsMegagroup              bool       `json:"isMegagroup"`                                              // IsMegagroup 表示超级群。
	IsBroadcast              bool       `json:"isBroadcast"`                                              // IsBroadcast 表示广播频道。
	IsArchived               bool       `json:"isArchived"`                                               // IsArchived 表示会话已归档。
	IsMuted                  bool       `json:"isMuted"`                                                  // IsMuted 表示会话已静音。
	UnreadCount              int        `json:"unreadCount"`                                              // UnreadCount 是未读消息数。
	LastMessageID            *int64     `json:"lastMessageId"`                                            // LastMessageID 是最后消息 ID。
	LastMessageAt            *time.Time `json:"lastMessageAt"`                                            // LastMessageAt 是最后消息时间。
	LastReadMessageID        *int64     `json:"lastReadMessageId"`                                        // LastReadMessageID 是最后已读消息 ID。
	Status                   string     `json:"status" gorm:"not null;default:normal"`                    // Status 是 Telegram 会话状态。
	IsCreator                bool       `json:"isCreator"`                                                // IsCreator 表示当前账号是创建者。
	ResourceComplianceStatus string     `json:"resourceComplianceStatus" gorm:"not null;default:allowed"` // ResourceComplianceStatus 是来源合规状态。
	CreatedAt                time.Time  `json:"CreatedAt"`
	UpdatedAt                time.Time  `json:"UpdatedAt"`
}

// TableName 返回 Telegram 来源表名。
func (TelegramDialog) TableName() string { return "tel_dialogs" }

// TelegramResourceTask 保存 Go 管理的 Telegram 批量采集任务。
type TelegramResourceTask struct {
	ID              uint       `json:"id" gorm:"primaryKey"`                                                                            // ID 是采集任务 ID。
	OwnerID         uint       `json:"-" gorm:"not null;index"`                                                                         // OwnerID 是任务所属用户。
	Keyword         string     `json:"keyword" gorm:"not null"`                                                                         // Keyword 是消息检索词。
	MaxDownloads    int        `json:"maxDownloads" gorm:"not null"`                                                                    // MaxDownloads 是文件数量上限。
	FileTypes       []string   `json:"fileTypes" gorm:"column:file_types_json;serializer:json;type:jsonb;not null;default:'[]'"`        // FileTypes 是允许的文件扩展名。
	Status          string     `json:"status" gorm:"not null;default:queued"`                                                           // Status 是任务状态。
	FilesFound      int        `json:"filesFound"`                                                                                      // FilesFound 是符合条件的文件数。
	FilesCompleted  int        `json:"filesCompleted"`                                                                                  // FilesCompleted 是成功登记的文件数。
	ErrorMessage    string     `json:"errorMessage"`                                                                                    // ErrorMessage 是安全错误说明。
	StartedAt       *time.Time `json:"startedAt"`                                                                                       // StartedAt 是开始时间。
	CompletedAt     *time.Time `json:"completedAt"`                                                                                     // CompletedAt 是结束时间。
	CancelRequested bool       `json:"cancelRequested"`                                                                                 // CancelRequested 表示用户请求取消。
	SourceMode      string     `json:"sourceMode" gorm:"not null"`                                                                      // SourceMode 是来源选择方式。
	SourceStrategy  string     `json:"sourceStrategy" gorm:"not null"`                                                                  // SourceStrategy 是实际选源策略。
	SourceDialogIDs []uint     `json:"sourceDialogIds" gorm:"column:source_selection;serializer:json;type:jsonb;not null;default:'[]'"` // SourceDialogIDs 是来源记录 ID。
	CreatedAt       time.Time  `json:"createdAt"`                                                                                       // CreatedAt 是创建时间。
	UpdatedAt       time.Time  `json:"updatedAt"`                                                                                       // UpdatedAt 是更新时间。
}

// TelegramResourceOptions 返回采集表单可选项。
type TelegramResourceOptions struct {
	FileTypes   []string `json:"fileTypes"`   // FileTypes 是支持的文件扩展名。
	MaxFileSize int64    `json:"maxFileSize"` // MaxFileSize 是单个文件的大小上限。
}

// TableName 返回 Telegram 采集任务表名。
func (TelegramResourceTask) TableName() string { return "tel_resource_tasks" }
