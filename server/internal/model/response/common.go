package response

// PageResult 是列表接口统一分页响应。
type PageResult struct {
	List     any   `json:"list"`     // List 是当前页业务记录集合。
	Total    int64 `json:"total"`    // Total 是符合条件的记录总数。
	Page     int   `json:"page"`     // Page 是当前页码。
	PageSize int   `json:"pageSize"` // PageSize 是当前每页数量。
}
