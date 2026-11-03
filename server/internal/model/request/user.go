package request

// Register 描述管理员创建系统用户的输入。
type Register struct {
	Username    string `json:"userName" example:"用户名"`   // Username 是唯一登录名。
	Password    string `json:"passWord" example:"密码"`    // Password 是待生成摘要的明文密码。
	NickName    string `json:"nickName" example:"昵称"`    // NickName 是用户显示名称。
	HeaderImage string `json:"headerImg" example:"头像链接"` // HeaderImage 是用户头像地址。
	AuthorityID uint   `json:"authorityId"`              // AuthorityID 是固定管理员或普通用户角色 ID。
	Enable      int    `json:"enable"`                   // Enable 表示用户状态：1 正常，2 冻结。
	Phone       string `json:"phone" example:"电话号码"`     // Phone 是用户手机号。
	Email       string `json:"email" example:"电子邮箱"`     // Email 是用户邮箱。
}

// Login 描述系统用户登录凭据。
type Login struct {
	Username string `json:"username"` // Username 是登录名。
	Password string `json:"password"` // Password 是明文密码，仅用于本次认证。
}

// ChangePasswordReq 描述当前用户修改自身密码的输入。
type ChangePasswordReq struct {
	ID          uint   `json:"-"`           // ID 从 JWT 中提取，不接受客户端指定。
	Password    string `json:"password"`    // Password 是当前明文密码。
	NewPassword string `json:"newPassword"` // NewPassword 是待设置的新明文密码。
}

// ResetPassword 描述管理员重置用户密码的输入。
type ResetPassword struct {
	ID       uint   `json:"ID" form:"ID"`                                   // ID 是目标系统用户 ID。
	Password string `json:"password" form:"password" gorm:"comment:用户登录密码"` // Password 是待设置的新明文密码。
}

// ChangeUserInfo 描述管理员可修改的用户资料与状态。
type ChangeUserInfo struct {
	ID          uint   `json:"ID"`          // ID 是目标系统用户 ID。
	NickName    string `json:"nickName"`    // NickName 是用户显示名称。
	Phone       string `json:"phone"`       // Phone 是用户手机号。
	AuthorityID uint   `json:"authorityId"` // AuthorityID 是固定管理员或普通用户角色 ID。
	Email       string `json:"email"`       // Email 是用户邮箱。
	HeaderImage string `json:"headerImg"`   // HeaderImage 是用户头像地址。
	Enable      int    `json:"enable"`      // Enable 表示用户状态：1 正常，2 冻结。
}

// UserSearch 描述系统用户列表的分页、筛选与排序条件。
type UserSearch struct {
	PageInfo        // PageInfo 提供分页和关键词条件。
	Username string `json:"username" form:"username"` // Username 按登录名筛选。
	NickName string `json:"nickName" form:"nickName"` // NickName 按显示名称筛选。
	Phone    string `json:"phone" form:"phone"`       // Phone 按手机号筛选。
	Email    string `json:"email" form:"email"`       // Email 按邮箱筛选。
	OrderKey string `json:"orderKey" form:"orderKey"` // OrderKey 是允许的排序字段名。
	Desc     bool   `json:"desc" form:"desc"`         // Desc 决定是否按降序排列。
}
