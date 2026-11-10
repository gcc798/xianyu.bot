package response

// DictionaryOption 是供前端选择和显示的字典项。
type DictionaryOption struct {
	Code string `json:"code"` // Code 是业务存储值。
	Name string `json:"name"` // Name 是界面显示名称。
}
