package response

// QRLogin 描述一次闲鱼二维码登录会话的当前状态。
type QRLogin struct {
	// SessionID 是二维码登录会话标识。
	SessionID string `json:"sessionId"`
	// AccountID 是登录成功后关联的闲鱼账号 ID。
	AccountID uint `json:"accountId"`
	// Status 是二维码生成、等待、成功或失败状态。
	Status string `json:"status"`
	// QRCodeDataURL 是可直接展示的二维码图片。
	QRCodeDataURL string `json:"qrCodeDataUrl,omitempty"`
	// Message 是当前状态的用户提示。
	Message string `json:"message"`
	// VerificationURL 是需要浏览器继续操作的验证地址。
	VerificationURL string `json:"verificationUrl,omitempty"`
}
