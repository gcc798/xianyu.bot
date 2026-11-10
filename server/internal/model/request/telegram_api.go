package request

// SaveTelegramAPIConfig 描述新增或更新 Telegram API 配置的输入。
type SaveTelegramAPIConfig struct {
	Name    string `json:"name" binding:"required"`  // Name 是配置名称。
	APIID   int64  `json:"apiId" binding:"required"` // APIID 是 Telegram 应用 ID。
	APIHash string `json:"apiHash"`                  // APIHash 是新凭据，空值表示保留现有凭据。
}

// TelegramAPIConfigSearch 描述 Telegram API 配置分页查询条件。
type TelegramAPIConfigSearch struct {
	Pagination
	Keyword string `form:"q"` // Keyword 匹配名称或 API ID。
}

// ReportTelegramFloodWait 描述 Telegram API 限流上报。
type ReportTelegramFloodWait struct {
	APIID   int64 `json:"apiId" binding:"required"`   // APIID 是触发限流的应用 ID。
	Seconds int64 `json:"seconds" binding:"required"` // Seconds 是 Telegram 返回的限流秒数。
}
