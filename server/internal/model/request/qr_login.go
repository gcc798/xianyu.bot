package request

// QRSessionID 标识一次二维码登录会话。
type QRSessionID struct {
	SessionID string `json:"sessionId" form:"sessionId" binding:"required"` // SessionID 是创建二维码时返回的会话标识。
}
