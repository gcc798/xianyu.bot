package service

import (
	"errors"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAgentSessionScopesConversationByOwner(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentSession{}, &model.AgentMessage{}, &model.AgentRun{}, &model.AgentRunTool{}))
	oldDB := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = oldDB })

	sessions := new(agentSession)
	owned, err := sessions.Create(t.Context(), 11, xyRequest.CreateAgentSession{})
	require.NoError(t, err)
	require.Equal(t, "新对话", owned.Title)
	other, err := sessions.Create(t.Context(), 22, xyRequest.CreateAgentSession{Title: "私有会话"})
	require.NoError(t, err)

	now := time.Now()
	require.NoError(t, db.Model(&owned).Update("last_message_at", now).Error)
	require.NoError(t, db.Create(&[]model.AgentMessage{
		{OwnerID: 11, SessionID: owned.ID, Sequence: 2, Role: "assistant", Content: "第二条"},
		{OwnerID: 11, SessionID: owned.ID, Sequence: 1, Role: "user", Content: "第一条"},
		{OwnerID: 22, SessionID: other.ID, Sequence: 1, Role: "user", Content: "不可见"},
	}).Error)
	run := model.AgentRun{OwnerID: 11, SessionID: owned.ID, ModelID: 1, ModelName: "test", Status: "completed", ConfigSnapshot: `{}`, StartedAt: now}
	require.NoError(t, db.Create(&run).Error)
	require.NoError(t, db.Create(&model.AgentRunTool{RunID: run.ID, CallID: "approval-1", ToolName: "resource_acquire", Status: "waiting_approval", StartedAt: now}).Error)

	list, total, err := sessions.List(t.Context(), 11, xyRequest.AgentSessionSearch{})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, owned.ID, list[0].ID)

	detail, err := sessions.Get(t.Context(), 11, owned.ID)
	require.NoError(t, err)
	require.Len(t, detail.Messages, 2)
	require.Len(t, detail.PendingApprovals, 1)
	require.Equal(t, "resource_acquire", detail.PendingApprovals[0].ToolName)
	require.EqualValues(t, 1, detail.Messages[0].Sequence)
	_, err = sessions.Get(t.Context(), 11, other.ID)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestAgentSessionDeleteRejectsRunningAndRemovesConversation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentSession{}, &model.AgentMessage{}, &model.AgentRun{}, &model.AgentRunTool{}))
	oldDB := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = oldDB })

	session := model.AgentSession{OwnerID: 11, Title: "待删除"}
	require.NoError(t, db.Create(&session).Error)
	run := model.AgentRun{OwnerID: 11, SessionID: session.ID, ModelID: 1, ModelName: "test", Status: "running", StartedAt: time.Now()}
	require.NoError(t, db.Create(&run).Error)
	require.NoError(t, db.Create(&model.AgentMessage{OwnerID: 11, SessionID: session.ID, RunID: &run.ID, Sequence: 1, Role: "user", Content: "内容"}).Error)
	require.NoError(t, db.Create(&model.AgentRunTool{RunID: run.ID, CallID: "call-1", ToolName: "resource_search", Status: "completed", StartedAt: time.Now()}).Error)

	sessions := new(agentSession)
	require.ErrorContains(t, sessions.Delete(t.Context(), 11, session.ID), "正在运行")
	require.NoError(t, db.Model(&run).Update("status", "completed").Error)
	require.NoError(t, sessions.Delete(t.Context(), 11, session.ID))

	for _, table := range []any{&model.AgentSession{}, &model.AgentMessage{}, &model.AgentRun{}, &model.AgentRunTool{}} {
		var count int64
		require.NoError(t, db.Model(table).Count(&count).Error)
		require.Zero(t, count)
	}
}
