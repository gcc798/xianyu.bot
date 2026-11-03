package model

// BusinessEvent 记录闲鱼自动化过程中的可审计业务结果。
type BusinessEvent struct {
	// Timestamps 保存事件记录的生命周期字段。
	Timestamps
	// OwnerID 是事件所属的系统用户 ID。
	OwnerID uint `json:"-" gorm:"not null;index:idx_xy_business_events_owner"`
	// AccountID 是事件关联的闲鱼账号 ID。
	AccountID uint `json:"accountId" gorm:"index"`
	// ProductID 是事件关联的商品 ID。
	ProductID uint `json:"productId" gorm:"index"`
	// OrderID 是事件关联的订单 ID。
	OrderID uint `json:"orderId" gorm:"index"`
	// Type 标识业务事件类别。
	Type string `json:"type" gorm:"size:32;not null;index"`
	// Status 标识业务动作是否成功。
	Status string `json:"status" gorm:"size:16;not null;index"`
	// Summary 是供用户查看的事件摘要。
	Summary string `json:"summary" gorm:"size:255;not null"`
	// Error 是失败时的安全错误描述。
	Error string `json:"error" gorm:"size:500"`
}

// TableName 返回业务事件表名。
func (BusinessEvent) TableName() string { return "xy_business_events" }
