package request

// ResourceSearch 描述本地资源列表的分页与筛选条件。
type ResourceSearch struct {
	Pagination        // Pagination 提供页码与每页数量。
	Keyword    string `json:"keyword" form:"keyword"`   // Keyword 按资源名称检索。
	MIMEType   string `json:"mimeType" form:"mimeType"` // MIMEType 按媒体类型筛选。
	State      string `json:"state" form:"state"`       // State 按资源可用状态筛选。
}

// CompleteResourceDelivery 描述资源上传完成后的校验信息。
type CompleteResourceDelivery struct {
	Size   int64  `json:"size" binding:"required"`   // Size 是已上传对象的字节数。
	SHA256 string `json:"sha256" binding:"required"` // SHA256 是已上传对象的内容摘要。
}

// UpdateResourceDelivery 描述资源长任务的进度更新。
type UpdateResourceDelivery struct {
	Stage           string `json:"stage" binding:"required,max=64"`         // Stage 是当前处理阶段。
	ProgressPercent int    `json:"progressPercent" binding:"min=0,max=100"` // ProgressPercent 是零到一百的完成百分比。
	ProgressDone    int64  `json:"progressDone" binding:"min=0"`            // ProgressDone 是已完成工作量。
	ProgressTotal   int64  `json:"progressTotal" binding:"min=0"`           // ProgressTotal 是总工作量。
	ProgressUnit    string `json:"progressUnit" binding:"max=32"`           // ProgressUnit 是工作量单位。
}

// FailResourceDelivery 描述资源长任务的失败结果。
type FailResourceDelivery struct {
	ErrorCode    string `json:"errorCode" binding:"required,max=64"`      // ErrorCode 是供程序处理的稳定错误编码。
	ErrorMessage string `json:"errorMessage" binding:"required,max=2000"` // ErrorMessage 是供用户诊断的安全错误描述。
}
