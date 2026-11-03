package request

// AccountSearch 描述账号列表的分页与筛选条件。
type AccountSearch struct {
	PageInfo        // PageInfo 提供分页和关键词条件。
	Status   string `json:"status" form:"status"`   // Status 按账号状态筛选。
	Enabled  *bool  `json:"enabled" form:"enabled"` // Enabled 按启用状态筛选，空值表示不限。
}

// CreateAccount 描述新建闲鱼账号的可编辑信息。
type CreateAccount struct {
	Label string `json:"label"` // Label 是账号显示名称。
	Note  string `json:"note"`  // Note 是账号备注。
}

// UpdateAccount 描述账号基本信息与启用状态更新。
type UpdateAccount struct {
	ID      uint   `json:"ID"`      // ID 是待更新的账号 ID。
	Label   string `json:"label"`   // Label 是账号显示名称。
	Note    string `json:"note"`    // Note 是账号备注。
	Enabled bool   `json:"enabled"` // Enabled 决定账号是否参与自动化任务。
}

// AccountID 标识一个闲鱼账号。
type AccountID struct {
	ID uint `json:"ID" form:"ID"` // ID 是目标账号 ID。
}

// SaveCredential 描述需要加密保存的账号认证材料。
type SaveCredential struct {
	ID     uint   `json:"ID"`     // ID 是凭据所属的账号 ID。
	Cookie string `json:"cookie"` // Cookie 是待加密的闲鱼 Cookie。
	Token  string `json:"token"`  // Token 是待加密的闲鱼访问令牌。
}
