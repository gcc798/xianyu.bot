package agenttool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestCurrentTimeReturnsServerDate 验证时间工具返回准确日期、星期和时区。
func TestCurrentTimeReturnsServerDate(t *testing.T) {
	tool := &currentTimeTool{now: func() time.Time {
		return time.Date(2026, time.November, 9, 9, 30, 15, 0, time.FixedZone("Asia/Shanghai", 8*60*60))
	}}

	result, err := tool.Invoke(t.Context(), Scope{OwnerID: 11}, []byte(`{}`))
	require.NoError(t, err)
	require.JSONEq(t, `{"date":"2026-11-09","time":"09:30:15","weekday":"星期一","timezone":"Asia/Shanghai"}`, result)
	require.True(t, tool.Definition().ReadOnly)
	require.False(t, tool.Definition().RequiresApproval)
}
