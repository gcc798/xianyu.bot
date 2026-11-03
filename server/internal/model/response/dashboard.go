package response

// DashboardOverview 汇总指定时间范围内的经营与自动化指标。
type DashboardOverview struct {
	// StartTime 是实际统计区间起始时间戳。
	StartTime int64 `json:"startTime"`
	// EndTime 是实际统计区间结束时间戳。
	EndTime int64 `json:"endTime"`
	// BucketMillis 是趋势图时间桶宽度毫秒数。
	BucketMillis int64 `json:"bucketMillis"`
	// Metrics 是区间内的核心汇总指标。
	Metrics DashboardMetrics `json:"metrics"`
	// MessageTrend 是消息量时间序列。
	MessageTrend []DashboardMessagePoint `json:"messageTrend"`
	// MessageSources 是消息来源分布。
	MessageSources []DashboardNameValue `json:"messageSources"`
	// OrderTrend 是订单与发货时间序列。
	OrderTrend []DashboardOrderPoint `json:"orderTrend"`
	// Automation 是自动化任务结果汇总。
	Automation DashboardAutomation `json:"automation"`
	// RecentEvents 是最近业务事件。
	RecentEvents []BusinessEvent `json:"recentEvents"`
}

// DashboardMetrics 是仪表盘顶部核心指标。
type DashboardMetrics struct {
	// Accounts 是账号总数。
	Accounts int64 `json:"accounts"`
	// ReadyAccounts 是可用账号数。
	ReadyAccounts int64 `json:"readyAccounts"`
	// OnSaleProducts 是在售商品数。
	OnSaleProducts int64 `json:"onSaleProducts"`
	// InboundMessages 是收到的消息数。
	InboundMessages int64 `json:"inboundMessages"`
	// OutboundMessages 是发出的消息数。
	OutboundMessages int64 `json:"outboundMessages"`
	// AutoReplies 是自动回复消息数。
	AutoReplies int64 `json:"autoReplies"`
	// AutoReplyRate 是入站消息的自动回复比例。
	AutoReplyRate float64 `json:"autoReplyRate"`
	// NewOrders 是新增订单数。
	NewOrders int64 `json:"newOrders"`
	// PendingOrders 是待发货订单数。
	PendingOrders int64 `json:"pendingOrders"`
	// FailedAutomationTasks 是失败自动化任务数。
	FailedAutomationTasks int64 `json:"failedAutomationTasks"`
}

// DashboardMessagePoint 是单个时间桶的消息统计。
type DashboardMessagePoint struct {
	// Timestamp 是时间桶起始时间戳。
	Timestamp int64 `json:"timestamp"`
	// Inbound 是收到的消息数。
	Inbound int64 `json:"inbound"`
	// Outbound 是发出的消息数。
	Outbound int64 `json:"outbound"`
	// AutoReplies 是自动回复消息数。
	AutoReplies int64 `json:"autoReplies"`
}

// DashboardOrderPoint 是单个时间桶的订单与发货统计。
type DashboardOrderPoint struct {
	// Timestamp 是时间桶起始时间戳。
	Timestamp int64 `json:"timestamp"`
	// Orders 是新增订单数。
	Orders int64 `json:"orders"`
	// Delivered 是成功发货数。
	Delivered int64 `json:"delivered"`
	// Failed 是失败发货数。
	Failed int64 `json:"failed"`
}

// DashboardNameValue 表示一个命名分类及其数量。
type DashboardNameValue struct {
	// Name 是分类名称。
	Name string `json:"name"`
	// Value 是分类数量。
	Value int64 `json:"value"`
}

// DashboardAutomation 汇总自动回复、发货和重发结果。
type DashboardAutomation struct {
	// ReplySuccess 是成功自动回复数。
	ReplySuccess int64 `json:"replySuccess"`
	// ReplyFailed 是失败自动回复数。
	ReplyFailed int64 `json:"replyFailed"`
	// DeliverySuccess 是成功自动发货数。
	DeliverySuccess int64 `json:"deliverySuccess"`
	// DeliveryFailed 是失败自动发货数。
	DeliveryFailed int64 `json:"deliveryFailed"`
	// RelistSuccess 是成功自动重发数。
	RelistSuccess int64 `json:"relistSuccess"`
	// RelistFailed 是失败自动重发数。
	RelistFailed int64 `json:"relistFailed"`
}
