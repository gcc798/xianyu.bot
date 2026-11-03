package model

// DictionaryDetail 表示字典中的一个可分层值条目。
type DictionaryDetail struct {
	// Timestamps 保存字典条目的生命周期字段。
	Timestamps
	// Label 是供界面展示的值。
	Label string `json:"label" form:"label" gorm:"column:label;comment:展示值"`
	// Value 是业务使用的字典值。
	Value string `json:"value" form:"value" gorm:"column:value;comment:字典值"`
	// Extend 保存可选扩展值。
	Extend string `json:"extend" form:"extend" gorm:"column:extend;comment:扩展值"`
	// Status 表示条目是否启用。
	Status *bool `json:"status" form:"status" gorm:"column:status;comment:启用状态"`
	// Sort 是同级条目的排序值。
	Sort int `json:"sort" form:"sort" gorm:"column:sort;comment:排序标记"`
	// DictionaryID 是所属字典 ID。
	DictionaryID int `json:"sysDictionaryID" form:"sysDictionaryID" gorm:"column:sys_dictionary_id;comment:关联标记"`
	// ParentID 是上级条目 ID，空值表示根节点。
	ParentID *uint `json:"parentID" form:"parentID" gorm:"column:parent_id;comment:父级字典详情ID"`
	// Children 是当前条目的直接子节点。
	Children []DictionaryDetail `json:"children" gorm:"foreignKey:ParentID"`
	// Level 是从零开始的层级深度。
	Level int `json:"level" form:"level" gorm:"column:level;comment:层级深度"`
	// Path 是逗号分隔的祖先 ID 路径。
	Path string `json:"path" form:"path" gorm:"column:path;comment:层级路径"`
}

// TableName 返回字典条目表名。
func (DictionaryDetail) TableName() string {
	return "sys_dictionary_details"
}
