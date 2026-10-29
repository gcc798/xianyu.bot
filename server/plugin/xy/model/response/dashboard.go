package response

type DashboardOverview struct {
	StartTime      int64                   `json:"startTime"`
	EndTime        int64                   `json:"endTime"`
	BucketMillis   int64                   `json:"bucketMillis"`
	Metrics        DashboardMetrics        `json:"metrics"`
	MessageTrend   []DashboardMessagePoint `json:"messageTrend"`
	MessageSources []DashboardNameValue    `json:"messageSources"`
	OrderTrend     []DashboardOrderPoint   `json:"orderTrend"`
	Automation     DashboardAutomation     `json:"automation"`
	RecentEvents   []BusinessEvent         `json:"recentEvents"`
}

type DashboardMetrics struct {
	Accounts              int64   `json:"accounts"`
	ReadyAccounts         int64   `json:"readyAccounts"`
	OnSaleProducts        int64   `json:"onSaleProducts"`
	InboundMessages       int64   `json:"inboundMessages"`
	OutboundMessages      int64   `json:"outboundMessages"`
	AutoReplies           int64   `json:"autoReplies"`
	AutoReplyRate         float64 `json:"autoReplyRate"`
	NewOrders             int64   `json:"newOrders"`
	PendingOrders         int64   `json:"pendingOrders"`
	FailedAutomationTasks int64   `json:"failedAutomationTasks"`
}

type DashboardMessagePoint struct {
	Timestamp   int64 `json:"timestamp"`
	Inbound     int64 `json:"inbound"`
	Outbound    int64 `json:"outbound"`
	AutoReplies int64 `json:"autoReplies"`
}

type DashboardOrderPoint struct {
	Timestamp int64 `json:"timestamp"`
	Orders    int64 `json:"orders"`
	Delivered int64 `json:"delivered"`
	Failed    int64 `json:"failed"`
}

type DashboardNameValue struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type DashboardAutomation struct {
	ReplySuccess    int64 `json:"replySuccess"`
	ReplyFailed     int64 `json:"replyFailed"`
	DeliverySuccess int64 `json:"deliverySuccess"`
	DeliveryFailed  int64 `json:"deliveryFailed"`
	RelistSuccess   int64 `json:"relistSuccess"`
	RelistFailed    int64 `json:"relistFailed"`
}
