package response

import "time"

// Message 展示一条带账号和接管状态的闲鱼消息。
type Message struct {
	// ID 是消息记录 ID。
	ID uint `json:"ID"`
	// CreatedAt 是消息入库时间。
	CreatedAt time.Time `json:"CreatedAt"`
	// AccountID 是消息所属的闲鱼账号 ID。
	AccountID uint `json:"accountId"`
	// AccountLabel 是账号显示名称。
	AccountLabel string `json:"accountLabel"`
	// PNMID 是闲鱼侧消息去重标识。
	PNMID string `json:"pnmId"`
	// SessionID 是闲鱼会话标识。
	SessionID string `json:"sessionId"`
	// ContentType 是闲鱼消息内容类型编码。
	ContentType int `json:"contentType"`
	// Content 是消息文本或序列化内容。
	Content string `json:"content"`
	// SenderUserName 是发送方显示名称。
	SenderUserName string `json:"senderUserName"`
	// SenderUserID 是闲鱼侧发送方标识。
	SenderUserID string `json:"senderUserId"`
	// ExternalProductID 是消息关联的闲鱼商品标识。
	ExternalProductID string `json:"externalProductId"`
	// ExternalOrderID 是消息关联的闲鱼订单标识。
	ExternalOrderID string `json:"externalOrderId"`
	// MessageTime 是闲鱼侧消息时间戳。
	MessageTime int64 `json:"messageTime"`
	// Direction 标识消息是接收还是发送。
	Direction string `json:"direction"`
	// Source 标识消息来自远端同步或本地发送。
	Source string `json:"source"`
	// TakeoverActive 表示当前会话是否由用户接管。
	TakeoverActive bool `json:"takeoverActive"`
}
