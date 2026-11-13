package agenttool

import (
	"context"
	"encoding/json"
	"time"

	"github.com/eino-contrib/jsonschema"
)

// currentTimeTool 提供服务端可信时间。
type currentTimeTool struct {
	now func() time.Time // now 返回本次调用使用的当前时间。
}

// NewCurrentTime 创建只读时间工具。
func NewCurrentTime() Tool {
	return &currentTimeTool{now: time.Now}
}

// Definition 返回时间工具定义。
func (*currentTimeTool) Definition() Definition {
	return Definition{
		Code: "current_time", Name: "获取当前时间",
		Description: "仅在需要执行当刻的准确时间时调用。日期和星期使用本轮上下文。",
		Source:      "local", Parameters: jsonschema.Reflect(struct{}{}), ReadOnly: true,
		Timeout: time.Second, BudgetClass: BudgetGeneral,
	}
}

// Invoke 返回服务端当前时间的结构化结果。
func (t *currentTimeTool) Invoke(context.Context, Scope, json.RawMessage) (string, error) {
	now := t.now()
	weekdays := [...]string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}
	result, err := json.Marshal(struct {
		Date     string `json:"date"`     // Date 是服务端当前日期。
		Time     string `json:"time"`     // Time 是服务端当前时间。
		Weekday  string `json:"weekday"`  // Weekday 是服务端当前星期。
		Timezone string `json:"timezone"` // Timezone 是服务端时区。
	}{
		Date: now.Format(time.DateOnly), Time: now.Format(time.TimeOnly),
		Weekday: weekdays[now.Weekday()], Timezone: now.Location().String(),
	})
	if err != nil {
		return "", &ExecutionError{Code: "invalid_result", Message: "当前时间结果无效", Cause: err}
	}
	return string(result), nil
}
