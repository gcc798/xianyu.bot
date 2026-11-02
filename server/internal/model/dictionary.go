package model

type Dictionary struct {
	Timestamps
	Name     string             `json:"name" form:"name" gorm:"column:name;comment:字典名（中）"`              // 字典名（中）
	Type     string             `json:"type" form:"type" gorm:"column:type;comment:字典名（英）"`              // 字典名（英）
	Status   *bool              `json:"status" form:"status" gorm:"column:status;comment:状态"`            // 状态
	Desc     string             `json:"desc" form:"desc" gorm:"column:desc;comment:描述"`                  // 描述
	ParentID *uint              `json:"parentID" form:"parentID" gorm:"column:parent_id;comment:父级字典ID"` // 父级字典ID
	Children []Dictionary       `json:"children" gorm:"foreignKey:ParentID"`                             // 子字典
	Details  []DictionaryDetail `json:"sysDictionaryDetails" form:"sysDictionaryDetails"`
}

func (Dictionary) TableName() string {
	return "sys_dictionaries"
}
