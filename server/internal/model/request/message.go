package request

// MessageSearch 描述闲鱼消息列表的分页与筛选条件。
type MessageSearch struct {
	Pagination               // Pagination 提供页码与每页数量。
	AccountID         uint   `json:"accountId" form:"accountId"`                 // AccountID 仅返回指定闲鱼账号的消息。
	ProductID         uint   `json:"productId" form:"productId"`                 // ProductID 仅返回指定商品的消息。
	SessionID         string `json:"sessionId" form:"sessionId"`                 // SessionID 仅返回指定会话的消息。
	Direction         string `json:"direction" form:"direction"`                 // Direction 仅返回指定收发方向的消息。
	ExternalProductID string `json:"externalProductId" form:"externalProductId"` // ExternalProductID 仅返回指定闲鱼商品的消息。
	BuyerName         string `json:"buyerName" form:"buyerName"`                 // BuyerName 按买家显示名称筛选。
	Content           string `json:"content" form:"content"`                     // Content 按消息内容筛选。
	StartTime         int64  `json:"startTime" form:"startTime"`                 // StartTime 是消息时间范围起点。
	EndTime           int64  `json:"endTime" form:"endTime"`                     // EndTime 是消息时间范围终点。
}

// SendMessage 描述基于已有消息向同一会话回复的输入。
type SendMessage struct {
	MessageID uint   `json:"messageId"` // MessageID 是用于定位账号、会话和接收方的消息 ID。
	Text      string `json:"text"`      // Text 是发送给买家的文本。
}

// UpdateTakeover 描述会话人工接管状态更新。
type UpdateTakeover struct {
	MessageID uint `json:"messageId"` // MessageID 是用于定位目标会话的消息 ID。
	Active    bool `json:"active"`    // Active 决定是否启用人工接管。
}
