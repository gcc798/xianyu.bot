package response

// LoginStatus 描述闲鱼账号登录结果与凭据变化。
type LoginStatus struct {
	Status            string `json:"status"`            // Status 是账号当前登录状态。
	CookieUpdated     bool   `json:"cookieUpdated"`     // CookieUpdated 表示本次登录是否更新了 Cookie。
	CredentialVersion uint   `json:"credentialVersion"` // CredentialVersion 是当前凭据格式版本。
}
