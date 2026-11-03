package request

// Register 描述管理员创建系统用户的输入。
type Register struct {
	// Username 是唯一登录名。
	Username string `json:"userName" example:"用户名"`
	// Password 是待生成摘要的明文密码。
	Password string `json:"passWord" example:"密码"`
	// NickName 是用户显示名称。
	NickName string `json:"nickName" example:"昵称"`
	// HeaderImage 是用户头像地址。
	HeaderImage string `json:"headerImg" example:"头像链接"`
	// AuthorityID 是固定管理员或普通用户角色 ID。
	AuthorityID uint `json:"authorityId"`
	// Enable 表示用户状态：1 正常，2 冻结。
	Enable int `json:"enable"`
	// Phone 是用户手机号。
	Phone string `json:"phone" example:"电话号码"`
	// Email 是用户邮箱。
	Email string `json:"email" example:"电子邮箱"`
}

// Login 描述系统用户登录凭据。
type Login struct {
	// Username 是登录名。
	Username string `json:"username"`
	// Password 是明文密码，仅用于本次认证。
	Password string `json:"password"`
}

// ChangePasswordReq 描述当前用户修改自身密码的输入。
type ChangePasswordReq struct {
	// ID 从 JWT 中提取，不接受客户端指定。
	ID uint `json:"-"`
	// Password 是当前明文密码。
	Password string `json:"password"`
	// NewPassword 是待设置的新明文密码。
	NewPassword string `json:"newPassword"`
}

// ResetPassword 描述管理员重置用户密码的输入。
type ResetPassword struct {
	// ID 是目标系统用户 ID。
	ID uint `json:"ID" form:"ID"`
	// Password 是待设置的新明文密码。
	Password string `json:"password" form:"password" gorm:"comment:用户登录密码"`
}

// ChangeUserInfo 描述管理员可修改的用户资料与状态。
type ChangeUserInfo struct {
	// ID 是目标系统用户 ID。
	ID uint `json:"ID"`
	// NickName 是用户显示名称。
	NickName string `json:"nickName"`
	// Phone 是用户手机号。
	Phone string `json:"phone"`
	// AuthorityID 是固定管理员或普通用户角色 ID。
	AuthorityID uint `json:"authorityId"`
	// Email 是用户邮箱。
	Email string `json:"email"`
	// HeaderImage 是用户头像地址。
	HeaderImage string `json:"headerImg"`
	// Enable 表示用户状态：1 正常，2 冻结。
	Enable int `json:"enable"`
}

// UserSearch 描述系统用户列表的分页、筛选与排序条件。
type UserSearch struct {
	// PageInfo 提供分页和关键词条件。
	PageInfo
	// Username 按登录名筛选。
	Username string `json:"username" form:"username"`
	// NickName 按显示名称筛选。
	NickName string `json:"nickName" form:"nickName"`
	// Phone 按手机号筛选。
	Phone string `json:"phone" form:"phone"`
	// Email 按邮箱筛选。
	Email string `json:"email" form:"email"`
	// OrderKey 是允许的排序字段名。
	OrderKey string `json:"orderKey" form:"orderKey"`
	// Desc 决定是否按降序排列。
	Desc bool `json:"desc" form:"desc"`
}
