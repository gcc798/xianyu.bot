package agenttool

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eino-contrib/jsonschema"
)

// BudgetClass 标识工具调用消耗的 Agent 预算类别。
type BudgetClass string

const (
	// BudgetGeneral 使用通用工具调用预算。
	BudgetGeneral BudgetClass = "general"
	// BudgetSearch 同时消耗搜索调用预算。
	BudgetSearch BudgetClass = "search"
)

// Definition 描述工具稳定身份、模型契约和执行治理规则。
type Definition struct {
	Code             string             // Code 是持久化和审计使用的稳定编码。
	Name             string             // Name 是提供给模型的工具名称。
	Description      string             // Description 说明工具能力与边界。
	Source           string             // Source 标识 local 或 mcp 来源。
	ProviderCode     string             // ProviderCode 是远程工具所属 MCP 服务编码。
	Parameters       *jsonschema.Schema // Parameters 是模型可见的 JSON Schema。
	ReadOnly         bool               // ReadOnly 表示工具不会改变外部状态。
	RequiresApproval bool               // RequiresApproval 表示执行前必须获得用户确认。
	Timeout          time.Duration      // Timeout 是单次调用最长执行时间。
	BudgetClass      BudgetClass        // BudgetClass 决定消耗的运行预算。
}

// ExecutionError 是可安全返回给 Agent 的稳定工具错误。
type ExecutionError struct {
	Code    string // Code 是供程序处理的稳定错误编码。
	Message string // Message 是不含敏感信息的错误说明。
	Cause   error  // Cause 保留内部错误链供日志诊断。
}

// Error 返回安全消息并附加内部错误文本。
func (e *ExecutionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap 返回内部错误以支持错误链匹配。
func (e *ExecutionError) Unwrap() error { return e.Cause }

// Scope 是一次工具调用不可绕过的业务归属边界。
type Scope struct {
	OwnerID uint // OwnerID 是当前系统用户 ID。
}

// Tool 是 Agent 可调用能力的统一契约。
type Tool interface {
	// Definition 返回工具不可变定义。
	Definition() Definition
	// Invoke 在 scope 边界内使用 JSON 参数执行工具。
	Invoke(context.Context, Scope, json.RawMessage) (string, error)
}
