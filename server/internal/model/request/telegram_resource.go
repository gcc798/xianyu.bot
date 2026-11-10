package request

// CreateTelegramResourceTask 描述 Telegram 资源采集请求。
type CreateTelegramResourceTask struct {
	Keyword         string   `json:"keyword" binding:"required,min=2,max=200"`             // Keyword 是 Telegram 消息搜索词。
	FileTypes       []string `json:"fileTypes"`                                            // FileTypes 是允许采集的扩展名。
	SourceMode      string   `json:"sourceMode" binding:"required,oneof=smart all manual"` // SourceMode 是选源方式。
	SourceDialogIDs []uint   `json:"sourceDialogIds"`                                      // SourceDialogIDs 是手动选择的来源。
	MaxDownloads    int      `json:"maxDownloads" binding:"required,min=1,max=100"`        // MaxDownloads 是文件数量上限。
}
