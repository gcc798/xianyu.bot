package request

type QRSessionID struct {
	SessionID string `json:"sessionId" form:"sessionId" binding:"required"`
}
