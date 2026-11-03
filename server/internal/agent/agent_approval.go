package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Runner) Approve(ctx context.Context, ownerID, runID uint) (model.AgentRunTool, error) {
	if ownerID == 0 || runID == 0 {
		return model.AgentRunTool{}, errors.New("运行标识不能为空")
	}
	var run model.AgentRun
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", runID, ownerID).First(&run).Error; err != nil {
		return model.AgentRunTool{}, err
	}
	var calls []model.AgentRunTool
	if err := s.db.WithContext(ctx).Where("run_id = ? AND status = ?", runID, "waiting_approval").Order("id").Find(&calls).Error; err != nil {
		return model.AgentRunTool{}, err
	}
	if len(calls) == 0 {
		return model.AgentRunTool{}, errors.New("当前运行没有待确认操作")
	}
	call := calls[0]
	claimed := s.db.WithContext(ctx).Model(&call).Where("status = ?", "waiting_approval").Updates(map[string]any{
		"status": "approving", "error_code": "", "error_message": "",
	})
	if claimed.Error != nil {
		return model.AgentRunTool{}, claimed.Error
	}
	if claimed.RowsAffected != 1 {
		return model.AgentRunTool{}, errors.New("待确认操作正在处理")
	}

	if err := s.executeApproval(ctx, run, &call); err != nil {
		finishErr := s.finishApproval(ctx, &call, "failed", "approval_execution_failed", "确认后的操作执行失败")
		s.publish(run.SessionID, "tool.failed", map[string]any{
			"runId": run.ID, "callId": call.CallID, "server": call.MCPServerCode, "tool": call.ToolName,
			"error": map[string]any{"code": "approval_execution_failed", "message": "确认后的操作执行失败"},
		})
		return call, errors.Join(err, finishErr)
	}
	if err := s.finishApproval(ctx, &call, "completed", "", ""); err != nil {
		return call, err
	}
	s.publish(run.SessionID, "tool.completed", map[string]any{
		"runId": run.ID, "callId": call.CallID, "server": call.MCPServerCode, "tool": call.ToolName,
		"status": "completed", "durationMillis": time.Since(call.StartedAt).Milliseconds(),
		"resultRef": map[string]any{"type": "agent_run_tool", "id": call.ID},
	})
	return call, nil
}

func (s *Runner) executeApproval(ctx context.Context, run model.AgentRun, call *model.AgentRunTool) error {
	connections, err := s.mcpRegistryOrDefault().Discover(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = closeMCPConnections(connections) }()
	connection, definition := approvedTool(connections, call.MCPServerCode, call.ToolName)
	if connection == nil {
		return errors.New("MCP 能力暂不可用")
	}
	var input map[string]any
	if err := json.Unmarshal([]byte(call.Arguments), &input); err != nil {
		return errors.New("待确认参数无效")
	}
	var task *model.ResourceTask
	if call.ToolName == "resource_acquire" {
		task, err = s.prepareResourceApproval(ctx, run.OwnerID, call, input)
		if err != nil {
			return err
		}
	}
	requestContext, cancel := context.WithTimeout(ctx, min(mcpToolTimeout(definition), time.Duration(connection.Server.RequestTimeoutSeconds)*time.Second))
	defer cancel()
	result, err := connection.Client.CallTool(requestContext, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: call.ToolName, Arguments: input}})
	if err != nil {
		s.failApprovedResourceTask(ctx, task)
		return errors.New("MCP 能力暂不可用")
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return errors.New("MCP 工具结果无效")
	}
	if err := s.db.WithContext(ctx).Model(call).Update("result", string(payload)).Error; err != nil {
		return err
	}
	resultValue := mcpToolResultValue(result)
	if result.IsError || !approvedResultSucceeded(resultValue) {
		s.failApprovedResourceTask(ctx, task)
		return errors.New("MCP 操作执行失败")
	}
	if task != nil {
		externalTaskID := approvedExternalTaskID(resultValue)
		if externalTaskID == "" {
			s.failApprovedResourceTask(ctx, task)
			return errors.New("MCP 操作未返回任务标识")
		}
		now := time.Now()
		if err := s.db.WithContext(ctx).Model(task).Updates(map[string]any{
			"external_task_id": externalTaskID, "state": "running", "stage": "submitted", "started_at": now,
		}).Error; err != nil {
			return err
		}
		s.publish(run.SessionID, "task.created", map[string]any{"runId": run.ID, "taskId": task.ID, "state": "running"})
	}
	return nil
}

func approvedResultSucceeded(value any) bool {
	payload, _ := json.Marshal(value)
	var result struct {
		Status string `json:"status"`
	}
	_ = json.Unmarshal(payload, &result)
	return result.Status == "" || result.Status == "success"
}

func (s *Runner) failApprovedResourceTask(ctx context.Context, task *model.ResourceTask) {
	if task == nil {
		return
	}
	now := time.Now()
	_ = s.db.WithContext(ctx).Model(task).Updates(map[string]any{
		"state": "failed", "stage": "submit_failed", "error_code": "capability_unavailable",
		"error_message": "MCP 能力暂不可用", "completed_at": now,
	}).Error
}

func approvedTool(connections []*mcpConnection, serverCode, toolName string) (*mcpConnection, mcp.Tool) {
	for _, connection := range connections {
		if connection.Server.Code != serverCode {
			continue
		}
		for _, definition := range connection.Tools {
			if definition.Name == toolName {
				return connection, definition
			}
		}
	}
	return nil, mcp.Tool{}
}

func (s *Runner) prepareResourceApproval(ctx context.Context, ownerID uint, call *model.AgentRunTool, input map[string]any) (*model.ResourceTask, error) {
	externalID, _ := input["candidate_id"].(string)
	if externalID == "" {
		return nil, errors.New("资源候选无效")
	}
	var candidate model.ResourceCandidate
	if err := s.db.WithContext(ctx).Table("xy_resource_candidates AS c").
		Joins("JOIN xy_resource_searches AS s ON s.id = c.search_id AND s.deleted_at IS NULL").
		Where("c.external_id = ? AND c.mcp_server_code = ? AND c.deleted_at IS NULL AND s.owner_id = ?", externalID, call.MCPServerCode, ownerID).
		Select("c.*").Order("c.id DESC").First(&candidate).Error; err != nil {
		return nil, err
	}
	idempotencyKey := fmt.Sprintf("agent-run-%d-tool-%d", call.RunID, call.ID)
	grant, err := s.resources.CreateDeliveryTask(ctx, ownerID, candidate.ID, idempotencyKey)
	if err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Model(&grant.Task).Update("agent_run_id", call.RunID).Error; err != nil {
		return nil, err
	}
	grant.Task.AgentRunID = new(call.RunID)
	input["task_id"] = grant.Task.ID
	input["idempotency_key"] = idempotencyKey
	input["delivery_token"] = grant.Token
	return &grant.Task, nil
}

func approvedExternalTaskID(value any) string {
	payload, _ := json.Marshal(value)
	var result struct {
		Data struct {
			TaskID string `json:"task_id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(payload, &result)
	return result.Data.TaskID
}

func (s *Runner) finishApproval(ctx context.Context, call *model.AgentRunTool, status, errorCode, errorMessage string) error {
	completedAt := time.Now()
	finishContext, cancel := agentFinishContext(ctx)
	defer cancel()
	err := s.db.WithContext(finishContext).Model(call).Updates(map[string]any{
		"status": status, "error_code": errorCode, "error_message": errorMessage, "completed_at": completedAt,
	}).Error
	if err == nil {
		call.Status, call.ErrorCode, call.ErrorMessage, call.CompletedAt = status, errorCode, errorMessage, &completedAt
	}
	return err
}
