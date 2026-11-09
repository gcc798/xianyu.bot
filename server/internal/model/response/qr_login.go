package response

// QRLogin 描述一次闲鱼二维码登录会话的当前状态。
type QRLogin struct {
	SessionID       string `json:"sessionId"`                 // SessionID 是二维码登录会话标识。
	AccountID       uint   `json:"accountId"`                 // AccountID 是登录成功后关联的闲鱼账号 ID。
	Status          string `json:"status"`                    // Status 是二维码生成、等待、成功或失败状态。
	QRCodeDataURL   string `json:"qrCodeDataUrl,omitempty"`   // QRCodeDataURL 是可直接展示的二维码图片。
	Message         string `json:"message"`                   // Message 是当前状态的用户提示。
	VerificationURL string `json:"verificationUrl,omitempty"` // VerificationURL 是需要浏览器继续操作的验证地址。
}
