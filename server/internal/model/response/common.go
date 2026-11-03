package response

// PageResult 是列表接口统一分页响应。
type PageResult struct {
	// List 是当前页业务记录集合。
	List any `json:"list"`
	// Total 是符合条件的记录总数。
	Total int64 `json:"total"`
	// Page 是当前页码。
	Page int `json:"page"`
	// PageSize 是当前每页数量。
	PageSize int `json:"pageSize"`
}
