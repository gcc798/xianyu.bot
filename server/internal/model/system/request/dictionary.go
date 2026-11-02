package request

type DictionarySearch struct {
	Name string `json:"name" form:"name" gorm:"column:name;comment:字典名（中）"` // 字典名（中）
}
