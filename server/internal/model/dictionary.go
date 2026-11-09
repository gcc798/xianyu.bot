package model

// Dictionary 表示可分层组织的一组系统字典。
type Dictionary struct {
	Timestamps                    // Timestamps 保存字典记录的生命周期字段。
	Name       string             `json:"name" form:"name" gorm:"column:name;comment:字典名（中）"`              // Name 是字典显示名称。
	Type       string             `json:"type" form:"type" gorm:"column:type;comment:字典名（英）"`              // Type 是字典类型编码。
	Status     *bool              `json:"status" form:"status" gorm:"column:status;comment:状态"`            // Status 表示字典是否启用。
	Desc       string             `json:"desc" form:"desc" gorm:"column:desc;comment:描述"`                  // Desc 是字典用途说明。
	ParentID   *uint              `json:"parentID" form:"parentID" gorm:"column:parent_id;comment:父级字典ID"` // ParentID 是上级字典 ID，空值表示根节点。
	Children   []Dictionary       `json:"children" gorm:"foreignKey:ParentID"`                             // Children 是当前字典的直接子节点。
	Details    []DictionaryDetail `json:"sysDictionaryDetails" form:"sysDictionaryDetails"`                // Details 是字典包含的值条目。
}

// TableName 返回系统字典表名。
func (Dictionary) TableName() string {
	return "s_dictionaries"
}
