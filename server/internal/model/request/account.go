package request

// AccountSearch 描述账号列表的分页与筛选条件。
type AccountSearch struct {
	// PageInfo 提供分页和关键词条件。
	PageInfo
	// Status 按账号状态筛选。
	Status string `json:"status" form:"status"`
	// Enabled 按启用状态筛选，空值表示不限。
	Enabled *bool `json:"enabled" form:"enabled"`
}

// CreateAccount 描述新建闲鱼账号的可编辑信息。
type CreateAccount struct {
	// Label 是账号显示名称。
	Label string `json:"label"`
	// Note 是账号备注。
	Note string `json:"note"`
}

// UpdateAccount 描述账号基本信息与启用状态更新。
type UpdateAccount struct {
	// ID 是待更新的账号 ID。
	ID uint `json:"ID"`
	// Label 是账号显示名称。
	Label string `json:"label"`
	// Note 是账号备注。
	Note string `json:"note"`
	// Enabled 决定账号是否参与自动化任务。
	Enabled bool `json:"enabled"`
}

// AccountID 标识一个闲鱼账号。
type AccountID struct {
	// ID 是目标账号 ID。
	ID uint `json:"ID" form:"ID"`
}

// SaveCredential 描述需要加密保存的账号认证材料。
type SaveCredential struct {
	// ID 是凭据所属的账号 ID。
	ID uint `json:"ID"`
	// Cookie 是待加密的闲鱼 Cookie。
	Cookie string `json:"cookie"`
	// Token 是待加密的闲鱼访问令牌。
	Token string `json:"token"`
}
