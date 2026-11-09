package response

import "time"

// Message 展示一条带账号和接管状态的闲鱼消息。
type Message struct {
	ID                uint      `json:"ID"`                // ID 是消息记录 ID。
	CreatedAt         time.Time `json:"CreatedAt"`         // CreatedAt 是消息入库时间。
	AccountID         uint      `json:"accountId"`         // AccountID 是消息所属的闲鱼账号 ID。
	AccountLabel      string    `json:"accountLabel"`      // AccountLabel 是账号显示名称。
	PNMID             string    `json:"pnmId"`             // PNMID 是闲鱼侧消息去重标识。
	SessionID         string    `json:"sessionId"`         // SessionID 是闲鱼会话标识。
	ContentType       int       `json:"contentType"`       // ContentType 是闲鱼消息内容类型编码。
	Content           string    `json:"content"`           // Content 是消息文本或序列化内容。
	SenderUserName    string    `json:"senderUserName"`    // SenderUserName 是发送方显示名称。
	SenderUserID      string    `json:"senderUserId"`      // SenderUserID 是闲鱼侧发送方标识。
	ExternalProductID string    `json:"externalProductId"` // ExternalProductID 是消息关联的闲鱼商品标识。
	ExternalOrderID   string    `json:"externalOrderId"`   // ExternalOrderID 是消息关联的闲鱼订单标识。
	MessageTime       int64     `json:"messageTime"`       // MessageTime 是闲鱼侧消息时间戳。
	Direction         string    `json:"direction"`         // Direction 标识消息是接收还是发送。
	Source            string    `json:"source"`            // Source 标识消息来自远端同步或本地发送。
	TakeoverActive    bool      `json:"takeoverActive"`    // TakeoverActive 表示当前会话是否由用户接管。
}
