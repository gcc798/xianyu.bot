package request

// DictionarySearch 描述系统字典检索条件。
type DictionarySearch struct {
	// Name 按字典显示名称检索。
	Name string `json:"name" form:"name" gorm:"column:name;comment:字典名（中）"`
}
