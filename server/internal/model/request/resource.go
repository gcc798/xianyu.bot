package request

// ResourceSearch 描述本地资源列表的分页与筛选条件。
type ResourceSearch struct {
	// Pagination 提供页码与每页数量。
	Pagination
	// Keyword 按资源名称检索。
	Keyword string `json:"keyword" form:"keyword"`
	// MIMEType 按媒体类型筛选。
	MIMEType string `json:"mimeType" form:"mimeType"`
	// State 按资源可用状态筛选。
	State string `json:"state" form:"state"`
}

// CompleteResourceDelivery 描述资源上传完成后的校验信息。
type CompleteResourceDelivery struct {
	// Size 是已上传对象的字节数。
	Size int64 `json:"size" binding:"required"`
	// SHA256 是已上传对象的内容摘要。
	SHA256 string `json:"sha256" binding:"required"`
}

// UpdateResourceDelivery 描述资源长任务的进度更新。
type UpdateResourceDelivery struct {
	// Stage 是当前处理阶段。
	Stage string `json:"stage" binding:"required,max=64"`
	// ProgressPercent 是零到一百的完成百分比。
	ProgressPercent int `json:"progressPercent" binding:"min=0,max=100"`
	// ProgressDone 是已完成工作量。
	ProgressDone int64 `json:"progressDone" binding:"min=0"`
	// ProgressTotal 是总工作量。
	ProgressTotal int64 `json:"progressTotal" binding:"min=0"`
	// ProgressUnit 是工作量单位。
	ProgressUnit string `json:"progressUnit" binding:"max=32"`
}

// FailResourceDelivery 描述资源长任务的失败结果。
type FailResourceDelivery struct {
	// ErrorCode 是供程序处理的稳定错误编码。
	ErrorCode string `json:"errorCode" binding:"required,max=64"`
	// ErrorMessage 是供用户诊断的安全错误描述。
	ErrorMessage string `json:"errorMessage" binding:"required,max=2000"`
}
