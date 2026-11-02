package response

type LoginStatus struct {
	Status            string `json:"status"`
	CookieUpdated     bool   `json:"cookieUpdated"`
	CredentialVersion uint   `json:"credentialVersion"`
}
