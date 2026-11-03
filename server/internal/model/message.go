package model

import (
	"time"
)

// Message 保存闲鱼会话中一条收发消息的快照。
type Message struct {
	// Timestamps 保存消息记录的生命周期字段。
	Timestamps
	// OwnerID 是消息所属的系统用户 ID。
	OwnerID uint `json:"-" gorm:"not null;index:idx_xy_messages_owner"`
	// AccountID 是消息所属的闲鱼账号 ID。
	AccountID uint `json:"accountId" gorm:"not null;uniqueIndex:idx_xy_messages_account_pnm"`
	// PNMID 是闲鱼侧消息去重标识。
	PNMID string `json:"pnmId" gorm:"size:255;not null;uniqueIndex:idx_xy_messages_account_pnm"`
	// SessionID 是闲鱼会话标识。
	SessionID string `json:"sessionId" gorm:"size:255;index:idx_xy_messages_session"`
	// ContentType 是闲鱼消息内容类型编码。
	ContentType int `json:"contentType" gorm:"not null;default:0"`
	// Content 是消息文本或序列化内容。
	Content string `json:"content" gorm:"type:text"`
	// SenderUserName 是发送方展示名称。
	SenderUserName string `json:"senderUserName" gorm:"size:255"`
	// SenderUserID 是闲鱼侧发送方标识。
	SenderUserID string `json:"senderUserId" gorm:"size:128"`
	// SenderAppVersion 是发送方客户端版本。
	SenderAppVersion string `json:"senderAppVersion" gorm:"size:64"`
	// SenderPlatform 是发送方客户端平台。
	SenderPlatform string `json:"senderPlatform" gorm:"size:64"`
	// ReminderURL 是消息携带的提醒链接。
	ReminderURL string `json:"reminderUrl" gorm:"size:2048"`
	// ExternalProductID 是消息关联的闲鱼商品标识。
	ExternalProductID string `json:"externalProductId" gorm:"size:128;index"`
	// ExternalOrderID 是消息关联的闲鱼订单标识。
	ExternalOrderID string `json:"externalOrderId" gorm:"size:128;index"`
	// MessageTime 是闲鱼侧消息时间戳。
	MessageTime int64 `json:"messageTime" gorm:"not null;default:0;index"`
	// Direction 标识消息是接收还是发送。
	Direction string `json:"direction" gorm:"size:16;not null;default:inbound"`
	// Source 标识消息来自远端同步或本地发送。
	Source string `json:"source" gorm:"size:32;not null;default:remote"`
}

// TableName 返回闲鱼消息表名。
func (Message) TableName() string { return "xy_messages" }

// ConversationTakeover 表示用户暂时接管一个闲鱼会话。
type ConversationTakeover struct {
	// Timestamps 保存接管记录的生命周期字段。
	Timestamps
	// OwnerID 是接管记录所属的系统用户 ID。
	OwnerID uint `json:"-" gorm:"not null;index:idx_xy_takeovers_owner"`
	// AccountID 是会话所属的闲鱼账号 ID。
	AccountID uint `json:"accountId" gorm:"not null;uniqueIndex:idx_xy_takeovers_account_session"`
	// SessionID 是被接管的闲鱼会话标识。
	SessionID string `json:"sessionId" gorm:"size:255;not null;uniqueIndex:idx_xy_takeovers_account_session"`
	// Active 表示接管是否仍然生效。
	Active bool `json:"active" gorm:"not null;default:true"`
}

// TableName 返回会话接管表名。
func (ConversationTakeover) TableName() string { return "xy_conversation_takeovers" }

const (
	// ReplyTaskScheduled 表示回复任务等待发送。
	ReplyTaskScheduled = "scheduled"
	// ReplyTaskCancelled 表示回复任务已取消。
	ReplyTaskCancelled = "cancelled"
	// ReplyTaskSending 表示回复正在发送。
	ReplyTaskSending = "sending"
	// ReplyTaskSent 表示回复已确认发送。
	ReplyTaskSent = "sent"
	// ReplyTaskFailed 表示回复明确发送失败。
	ReplyTaskFailed = "failed"
	// ReplyTaskUnknown 表示发送结果尚未确认。
	ReplyTaskUnknown = "unknown"
)

// ReplyTask 记录一条自动回复的调度与发送结果。
type ReplyTask struct {
	// Timestamps 保存任务记录的生命周期字段。
	Timestamps
	// OwnerID 是任务所属的系统用户 ID。
	OwnerID uint `json:"-" gorm:"not null;index:idx_xy_reply_tasks_owner"`
	// AccountID 是执行回复的闲鱼账号 ID。
	AccountID uint `json:"accountId" gorm:"not null;index"`
	// MessageID 是触发回复的本地消息 ID。
	MessageID uint `json:"messageId" gorm:"not null;uniqueIndex"`
	// ProductID 是消息关联的本地商品 ID。
	ProductID uint `json:"productId" gorm:"not null;index"`
	// ExternalProductID 是闲鱼侧商品标识。
	ExternalProductID string `json:"externalProductId" gorm:"size:128;not null;index"`
	// RuleID 是命中的自动回复规则 ID。
	RuleID uint `json:"ruleId" gorm:"not null"`
	// SessionID 是目标闲鱼会话标识。
	SessionID string `json:"sessionId" gorm:"size:255;not null;index"`
	// RecipientID 是闲鱼侧接收方标识。
	RecipientID string `json:"recipientId" gorm:"size:128;not null"`
	// ReplyText 是最终发送的回复文本。
	ReplyText string `json:"replyText" gorm:"type:text;not null"`
	// Status 是任务调度或发送状态。
	Status string `json:"status" gorm:"size:16;not null;index"`
	// Error 是最近一次失败原因。
	Error string `json:"error" gorm:"size:500"`
	// ScheduledAt 是任务允许开始发送的时间。
	ScheduledAt time.Time `json:"scheduledAt" gorm:"not null"`
	// CompletedAt 是任务完成时间。
	CompletedAt *time.Time `json:"completedAt"`
}

// TableName 返回自动回复任务表名。
func (ReplyTask) TableName() string { return "xy_reply_tasks" }
