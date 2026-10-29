package response

type QRLogin struct {
	SessionID       string `json:"sessionId"`
	AccountID       uint   `json:"accountId"`
	Status          string `json:"status"`
	QRCodeDataURL   string `json:"qrCodeDataUrl,omitempty"`
	Message         string `json:"message"`
	VerificationURL string `json:"verificationUrl,omitempty"`
}
