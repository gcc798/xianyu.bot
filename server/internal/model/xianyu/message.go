package model

import (
	"time"

	"github.com/gcc798/go.xianyu.assistant/server/internal/model/common"
)

type Message struct {
	common.Timestamps
	OwnerID           uint   `json:"-" gorm:"not null;index:idx_xy_messages_owner"`
	AccountID         uint   `json:"accountId" gorm:"not null;uniqueIndex:idx_xy_messages_account_pnm"`
	PNMID             string `json:"pnmId" gorm:"size:255;not null;uniqueIndex:idx_xy_messages_account_pnm"`
	SessionID         string `json:"sessionId" gorm:"size:255;index:idx_xy_messages_session"`
	ContentType       int    `json:"contentType" gorm:"not null;default:0"`
	Content           string `json:"content" gorm:"type:text"`
	SenderUserName    string `json:"senderUserName" gorm:"size:255"`
	SenderUserID      string `json:"senderUserId" gorm:"size:128"`
	SenderAppVersion  string `json:"senderAppVersion" gorm:"size:64"`
	SenderPlatform    string `json:"senderPlatform" gorm:"size:64"`
	ReminderURL       string `json:"reminderUrl" gorm:"size:2048"`
	ExternalProductID string `json:"externalProductId" gorm:"size:128;index"`
	ExternalOrderID   string `json:"externalOrderId" gorm:"size:128;index"`
	MessageTime       int64  `json:"messageTime" gorm:"not null;default:0;index"`
	Direction         string `json:"direction" gorm:"size:16;not null;default:inbound"`
	Source            string `json:"source" gorm:"size:32;not null;default:remote"`
}

func (Message) TableName() string { return "xy_messages" }

type ConversationTakeover struct {
	common.Timestamps
	OwnerID   uint   `json:"-" gorm:"not null;index:idx_xy_takeovers_owner"`
	AccountID uint   `json:"accountId" gorm:"not null;uniqueIndex:idx_xy_takeovers_account_session"`
	SessionID string `json:"sessionId" gorm:"size:255;not null;uniqueIndex:idx_xy_takeovers_account_session"`
	Active    bool   `json:"active" gorm:"not null;default:true"`
}

func (ConversationTakeover) TableName() string { return "xy_conversation_takeovers" }

const (
	ReplyTaskScheduled = "scheduled"
	ReplyTaskCancelled = "cancelled"
	ReplyTaskSending   = "sending"
	ReplyTaskSent      = "sent"
	ReplyTaskFailed    = "failed"
	ReplyTaskUnknown   = "unknown"
)

type ReplyTask struct {
	common.Timestamps
	OwnerID           uint       `json:"-" gorm:"not null;index:idx_xy_reply_tasks_owner"`
	AccountID         uint       `json:"accountId" gorm:"not null;index"`
	MessageID         uint       `json:"messageId" gorm:"not null;uniqueIndex"`
	ProductID         uint       `json:"productId" gorm:"not null;index"`
	ExternalProductID string     `json:"externalProductId" gorm:"size:128;not null;index"`
	RuleID            uint       `json:"ruleId" gorm:"not null"`
	SessionID         string     `json:"sessionId" gorm:"size:255;not null;index"`
	RecipientID       string     `json:"recipientId" gorm:"size:128;not null"`
	ReplyText         string     `json:"replyText" gorm:"type:text;not null"`
	Status            string     `json:"status" gorm:"size:16;not null;index"`
	Error             string     `json:"error" gorm:"size:500"`
	ScheduledAt       time.Time  `json:"scheduledAt" gorm:"not null"`
	CompletedAt       *time.Time `json:"completedAt"`
}

func (ReplyTask) TableName() string { return "xy_reply_tasks" }
