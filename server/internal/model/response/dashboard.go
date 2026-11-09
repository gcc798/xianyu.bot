package response

// DashboardOverview 汇总指定时间范围内的经营与自动化指标。
type DashboardOverview struct {
	StartTime      int64                   `json:"startTime"`      // StartTime 是实际统计区间起始时间戳。
	EndTime        int64                   `json:"endTime"`        // EndTime 是实际统计区间结束时间戳。
	BucketMillis   int64                   `json:"bucketMillis"`   // BucketMillis 是趋势图时间桶宽度毫秒数。
	Metrics        DashboardMetrics        `json:"metrics"`        // Metrics 是区间内的核心汇总指标。
	MessageTrend   []DashboardMessagePoint `json:"messageTrend"`   // MessageTrend 是消息量时间序列。
	MessageSources []DashboardNameValue    `json:"messageSources"` // MessageSources 是消息来源分布。
	OrderTrend     []DashboardOrderPoint   `json:"orderTrend"`     // OrderTrend 是订单与发货时间序列。
	Automation     DashboardAutomation     `json:"automation"`     // Automation 是自动化任务结果汇总。
	RecentEvents   []BusinessEvent         `json:"recentEvents"`   // RecentEvents 是最近业务事件。
}

// DashboardMetrics 是仪表盘顶部核心指标。
type DashboardMetrics struct {
	Accounts              int64   `json:"accounts"`              // Accounts 是账号总数。
	ReadyAccounts         int64   `json:"readyAccounts"`         // ReadyAccounts 是可用账号数。
	OnSaleProducts        int64   `json:"onSaleProducts"`        // OnSaleProducts 是在售商品数。
	InboundMessages       int64   `json:"inboundMessages"`       // InboundMessages 是收到的消息数。
	OutboundMessages      int64   `json:"outboundMessages"`      // OutboundMessages 是发出的消息数。
	AutoReplies           int64   `json:"autoReplies"`           // AutoReplies 是自动回复消息数。
	AutoReplyRate         float64 `json:"autoReplyRate"`         // AutoReplyRate 是入站消息的自动回复比例。
	NewOrders             int64   `json:"newOrders"`             // NewOrders 是新增订单数。
	PendingOrders         int64   `json:"pendingOrders"`         // PendingOrders 是待发货订单数。
	FailedAutomationTasks int64   `json:"failedAutomationTasks"` // FailedAutomationTasks 是失败自动化任务数。
}

// DashboardMessagePoint 是单个时间桶的消息统计。
type DashboardMessagePoint struct {
	Timestamp   int64 `json:"timestamp"`   // Timestamp 是时间桶起始时间戳。
	Inbound     int64 `json:"inbound"`     // Inbound 是收到的消息数。
	Outbound    int64 `json:"outbound"`    // Outbound 是发出的消息数。
	AutoReplies int64 `json:"autoReplies"` // AutoReplies 是自动回复消息数。
}

// DashboardOrderPoint 是单个时间桶的订单与发货统计。
type DashboardOrderPoint struct {
	Timestamp int64 `json:"timestamp"` // Timestamp 是时间桶起始时间戳。
	Orders    int64 `json:"orders"`    // Orders 是新增订单数。
	Delivered int64 `json:"delivered"` // Delivered 是成功发货数。
	Failed    int64 `json:"failed"`    // Failed 是失败发货数。
}

// DashboardNameValue 表示一个命名分类及其数量。
type DashboardNameValue struct {
	Name  string `json:"name"`  // Name 是分类名称。
	Value int64  `json:"value"` // Value 是分类数量。
}

// DashboardAutomation 汇总自动回复、发货和重发结果。
type DashboardAutomation struct {
	ReplySuccess    int64 `json:"replySuccess"`    // ReplySuccess 是成功自动回复数。
	ReplyFailed     int64 `json:"replyFailed"`     // ReplyFailed 是失败自动回复数。
	DeliverySuccess int64 `json:"deliverySuccess"` // DeliverySuccess 是成功自动发货数。
	DeliveryFailed  int64 `json:"deliveryFailed"`  // DeliveryFailed 是失败自动发货数。
	RelistSuccess   int64 `json:"relistSuccess"`   // RelistSuccess 是成功自动重发数。
	RelistFailed    int64 `json:"relistFailed"`    // RelistFailed 是失败自动重发数。
}
