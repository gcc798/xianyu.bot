package response

// WebSocketStatus 描述单个闲鱼账号的消息连接运行状态。
type WebSocketStatus struct {
	AccountID     uint   `json:"accountId"`     // AccountID 是闲鱼账号 ID。
	State         string `json:"state"`         // State 是连接状态机当前状态。
	QueueDepth    int    `json:"queueDepth"`    // QueueDepth 是待处理消息队列长度。
	Reconnects    uint64 `json:"reconnects"`    // Reconnects 是进程启动后的重连次数。
	Running       bool   `json:"running"`       // Running 表示账号连接协程是否运行。
	LastError     string `json:"lastError"`     // LastError 是最近一次连接错误。
	AccountStatus string `json:"accountStatus"` // AccountStatus 是持久化账号状态。
}
