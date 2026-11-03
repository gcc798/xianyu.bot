package request

// MessageSearch 描述闲鱼消息列表的分页与筛选条件。
type MessageSearch struct {
	// Pagination 提供页码与每页数量。
	Pagination
	// AccountID 仅返回指定闲鱼账号的消息。
	AccountID uint `json:"accountId" form:"accountId"`
	// SessionID 仅返回指定会话的消息。
	SessionID string `json:"sessionId" form:"sessionId"`
	// Direction 仅返回指定收发方向的消息。
	Direction string `json:"direction" form:"direction"`
	// ExternalProductID 仅返回指定闲鱼商品的消息。
	ExternalProductID string `json:"externalProductId" form:"externalProductId"`
	// BuyerName 按买家显示名称筛选。
	BuyerName string `json:"buyerName" form:"buyerName"`
	// Content 按消息内容筛选。
	Content string `json:"content" form:"content"`
	// StartTime 是消息时间范围起点。
	StartTime int64 `json:"startTime" form:"startTime"`
	// EndTime 是消息时间范围终点。
	EndTime int64 `json:"endTime" form:"endTime"`
}

// SendMessage 描述基于已有消息向同一会话回复的输入。
type SendMessage struct {
	// MessageID 是用于定位账号、会话和接收方的消息 ID。
	MessageID uint `json:"messageId"`
	// Text 是发送给买家的文本。
	Text string `json:"text"`
}

// UpdateTakeover 描述会话人工接管状态更新。
type UpdateTakeover struct {
	// MessageID 是用于定位目标会话的消息 ID。
	MessageID uint `json:"messageId"`
	// Active 决定是否启用人工接管。
	Active bool `json:"active"`
}
