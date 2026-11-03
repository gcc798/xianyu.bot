package response

// LoginStatus 描述闲鱼账号登录结果与凭据变化。
type LoginStatus struct {
	// Status 是账号当前登录状态。
	Status string `json:"status"`
	// CookieUpdated 表示本次登录是否更新了 Cookie。
	CookieUpdated bool `json:"cookieUpdated"`
	// CredentialVersion 是当前凭据格式版本。
	CredentialVersion uint `json:"credentialVersion"`
}
