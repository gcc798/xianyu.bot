package request

// DashboardOverview 描述仪表盘统计范围。
type DashboardOverview struct {
	AccountID uint  `form:"accountId"`                    // AccountID 仅统计指定闲鱼账号，零值表示全部账号。
	StartTime int64 `form:"startTime" binding:"required"` // StartTime 是统计区间起始时间戳。
	EndTime   int64 `form:"endTime" binding:"required"`   // EndTime 是统计区间结束时间戳。
}
