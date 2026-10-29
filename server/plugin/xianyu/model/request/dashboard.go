package request

type DashboardOverview struct {
	AccountID uint  `form:"accountId"`
	StartTime int64 `form:"startTime" binding:"required"`
	EndTime   int64 `form:"endTime" binding:"required"`
}
