package request

// ResourceSearch 描述本地资源列表的分页与筛选条件。
type ResourceSearch struct {
	Pagination        // Pagination 提供页码与每页数量。
	Keyword    string `json:"keyword" form:"keyword"`   // Keyword 按资源名称检索。
	MIMEType   string `json:"mimeType" form:"mimeType"` // MIMEType 按媒体类型筛选。
	State      string `json:"state" form:"state"`       // State 按资源可用状态筛选。
}

// ResourceTaskSearch 描述资源获取任务的分页与状态筛选条件。
type ResourceTaskSearch struct {
	Pagination        // Pagination 提供页码与每页数量。
	State      string `json:"state" form:"state"` // State 按任务状态筛选。
}

// RegisterTelegramResource 描述 Telegram 下载完成后登记的资源信息。
type RegisterTelegramResource struct {
	OwnerID         uint   `json:"ownerId" binding:"required"`             // OwnerID 是资源所属的系统用户 ID。
	Name            string `json:"name" binding:"required,max=500"`        // Name 是资源文件名。
	Size            int64  `json:"size" binding:"required,min=1"`          // Size 是资源大小，单位为字节。
	MIMEType        string `json:"mimeType" binding:"max=200"`             // MIMEType 是资源媒体类型。
	SHA256          string `json:"sha256" binding:"required,len=64"`       // SHA256 是资源内容摘要。
	ObjectKey       string `json:"objectKey" binding:"required,max=1000"`  // ObjectKey 是 Telegram 下载对象的存储键。
	SourceName      string `json:"sourceName" binding:"max=500"`           // SourceName 是 Telegram 来源名称。
	SourceReference string `json:"sourceReference" binding:"max=1000"`     // SourceReference 是 Telegram 来源引用。
	ExternalID      string `json:"externalId" binding:"required,max=1000"` // ExternalID 是 Telegram 消息的稳定标识。
}

// DeleteTelegramResource 描述 Telegram 文件删除请求。
type DeleteTelegramResource struct {
	OwnerID uint `json:"ownerId" binding:"required"` // OwnerID 是文件所属的系统用户 ID。
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
