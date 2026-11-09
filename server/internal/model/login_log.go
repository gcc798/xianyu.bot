package model

// LoginLog 记录一次系统用户登录尝试及结果。
type LoginLog struct {
	Timestamps          // Timestamps 保存登录日志的生命周期字段。
	Username     string `json:"username" gorm:"column:username;comment:用户名"`           // Username 是本次尝试使用的登录名。
	IP           string `json:"ip" gorm:"column:ip;comment:请求ip"`                      // IP 是客户端来源地址。
	Status       bool   `json:"status" gorm:"column:status;comment:登录状态"`              // Status 表示登录是否成功。
	ErrorMessage string `json:"errorMessage" gorm:"column:error_message;comment:错误信息"` // ErrorMessage 是登录失败原因。
	Agent        string `json:"agent" gorm:"column:agent;comment:代理"`                  // Agent 是客户端 User-Agent。
	UserID       uint   `json:"userId" gorm:"column:user_id;comment:用户id"`             // UserID 是成功匹配的系统用户 ID。
	User         User   `json:"user" gorm:"foreignKey:UserID"`                         // User 是关联的系统用户。
}

// TableName 返回登录审计表名。
func (LoginLog) TableName() string { return "s_login_logs" }
