package model

// BusinessEvent 记录闲鱼自动化过程中的可审计业务结果。
type BusinessEvent struct {
	Timestamps        // Timestamps 保存事件记录的生命周期字段。
	OwnerID    uint   `json:"-" gorm:"not null;index:idx_xy_business_events_owner"` // OwnerID 是事件所属的系统用户 ID。
	AccountID  uint   `json:"accountId" gorm:"index"`                               // AccountID 是事件关联的闲鱼账号 ID。
	ProductID  uint   `json:"productId" gorm:"index"`                               // ProductID 是事件关联的商品 ID。
	OrderID    uint   `json:"orderId" gorm:"index"`                                 // OrderID 是事件关联的订单 ID。
	Type       string `json:"type" gorm:"size:32;not null;index"`                   // Type 标识业务事件类别。
	Status     string `json:"status" gorm:"size:16;not null;index"`                 // Status 标识业务动作是否成功。
	Summary    string `json:"summary" gorm:"size:255;not null"`                     // Summary 是供用户查看的事件摘要。
	Error      string `json:"error" gorm:"size:500"`                                // Error 是失败时的安全错误描述。
}

// TableName 返回业务事件表名。
func (BusinessEvent) TableName() string { return "xy_business_events" }
