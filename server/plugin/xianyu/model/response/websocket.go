package response

type WebSocketStatus struct {
	AccountID  uint   `json:"accountId"`
	State      string `json:"state"`
	QueueDepth int    `json:"queueDepth"`
	Reconnects uint64 `json:"reconnects"`
	Running    bool   `json:"running"`
	LastError  string `json:"lastError"`
}
