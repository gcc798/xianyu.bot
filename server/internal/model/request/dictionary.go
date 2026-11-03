package request

// DictionarySearch 描述系统字典检索条件。
type DictionarySearch struct {
	Name string `json:"name" form:"name" gorm:"column:name;comment:字典名（中）"` // Name 按字典显示名称检索。
}
