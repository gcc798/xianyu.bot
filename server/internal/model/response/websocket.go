package response

// WebSocketStatus 描述单个闲鱼账号的消息连接运行状态。
type WebSocketStatus struct {
	// AccountID 是闲鱼账号 ID。
	AccountID uint `json:"accountId"`
	// State 是连接状态机当前状态。
	State string `json:"state"`
	// QueueDepth 是待处理消息队列长度。
	QueueDepth int `json:"queueDepth"`
	// Reconnects 是进程启动后的重连次数。
	Reconnects uint64 `json:"reconnects"`
	// Running 表示账号连接协程是否运行。
	Running bool `json:"running"`
	// LastError 是最近一次连接错误。
	LastError string `json:"lastError"`
	// AccountStatus 是持久化账号状态。
	AccountStatus string `json:"accountStatus"`
}
