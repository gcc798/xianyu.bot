package model

// LoginLog 记录一次系统用户登录尝试及结果。
type LoginLog struct {
	// Timestamps 保存登录日志的生命周期字段。
	Timestamps
	// Username 是本次尝试使用的登录名。
	Username string `json:"username" gorm:"column:username;comment:用户名"`
	// IP 是客户端来源地址。
	IP string `json:"ip" gorm:"column:ip;comment:请求ip"`
	// Status 表示登录是否成功。
	Status bool `json:"status" gorm:"column:status;comment:登录状态"`
	// ErrorMessage 是登录失败原因。
	ErrorMessage string `json:"errorMessage" gorm:"column:error_message;comment:错误信息"`
	// Agent 是客户端 User-Agent。
	Agent string `json:"agent" gorm:"column:agent;comment:代理"`
	// UserID 是成功匹配的系统用户 ID。
	UserID uint `json:"userId" gorm:"column:user_id;comment:用户id"`
	// User 是关联的系统用户。
	User User `json:"user" gorm:"foreignKey:UserID"`
}
