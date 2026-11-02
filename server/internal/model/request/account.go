package request

type AccountSearch struct {
	PageInfo
	Status  string `json:"status" form:"status"`
	Enabled *bool  `json:"enabled" form:"enabled"`
}

type CreateAccount struct {
	Label string `json:"label"`
	Note  string `json:"note"`
}

type UpdateAccount struct {
	ID      uint   `json:"ID"`
	Label   string `json:"label"`
	Note    string `json:"note"`
	Enabled bool   `json:"enabled"`
}

type AccountID struct {
	ID uint `json:"ID" form:"ID"`
}

type SaveCredential struct {
	ID     uint   `json:"ID"`
	Cookie string `json:"cookie"`
	Token  string `json:"token"`
}
