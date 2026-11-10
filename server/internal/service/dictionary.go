package service

import (
	"context"
	"errors"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"gorm.io/gorm"
)

// DictionaryService 管理可分层的系统字典。
type DictionaryService struct {
	db *gorm.DB // db 持久化字典及层级关系。
}

// NewDictionaryService 创建系统字典服务。
func NewDictionaryService(db *gorm.DB) *DictionaryService {
	return &DictionaryService{db: db}
}

// Create 在类型编码唯一时创建字典。
func (dictionaryService *DictionaryService) Create(dictionary model.Dictionary) error {
	if !errors.Is(dictionaryService.db.First(&model.Dictionary{}, "type = ?", dictionary.Type).Error, gorm.ErrRecordNotFound) {
		return errors.New("存在相同的type，不允许创建")
	}
	return dictionaryService.db.Create(&dictionary).Error
}

// Delete 删除字典及其直属条目；不存在时返回明确错误。
func (dictionaryService *DictionaryService) Delete(dictionary model.Dictionary) error {
	err := dictionaryService.db.Where("id = ?", dictionary.ID).Preload("Details").First(&dictionary).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("字典不存在")
	}
	if err != nil {
		return err
	}
	if err := dictionaryService.db.Delete(&dictionary).Error; err != nil {
		return err
	}
	if dictionary.Details != nil {
		return dictionaryService.db.Where("sys_dictionary_id = ?", dictionary.ID).Delete(dictionary.Details).Error
	}
	return nil
}

// Update 校验类型唯一性和父级环路后更新字典。
func (dictionaryService *DictionaryService) Update(dictionary *model.Dictionary) error {
	var stored model.Dictionary
	if err := dictionaryService.db.Where("id = ?", dictionary.ID).First(&stored).Error; err != nil {
		return errors.New("查询字典数据失败")
	}
	if stored.Type != dictionary.Type && !errors.Is(dictionaryService.db.First(&model.Dictionary{}, "type = ?", dictionary.Type).Error, gorm.ErrRecordNotFound) {
		return errors.New("存在相同的type，不允许创建")
	}
	if dictionary.ParentID != nil && *dictionary.ParentID != 0 {
		if err := dictionaryService.checkCircularReference(dictionary.ID, *dictionary.ParentID); err != nil {
			return err
		}
	}
	return dictionaryService.db.Model(&stored).Updates(map[string]any{
		"Name": dictionary.Name, "Type": dictionary.Type, "Status": dictionary.Status,
		"Desc": dictionary.Desc, "ParentID": dictionary.ParentID,
	}).Error
}

// List 按名称或类型检索字典并预加载直接子节点。
func (dictionaryService *DictionaryService) List(ctx context.Context, search request.DictionarySearch) ([]model.Dictionary, error) {
	var dictionaries []model.Dictionary
	query := dictionaryService.db.WithContext(ctx)
	if search.Name != "" {
		query = query.Where("name LIKE ? OR type LIKE ?", "%"+search.Name+"%", "%"+search.Name+"%")
	}
	err := query.Preload("Children").Find(&dictionaries).Error
	return dictionaries, err
}

// Options 返回所有启用字典项的业务编码与展示名称。
func (dictionaryService *DictionaryService) Options(ctx context.Context) (map[string][]response.DictionaryOption, error) {
	var rows []struct {
		Type string `gorm:"column:type"`
		Code string `gorm:"column:code"`
		Name string `gorm:"column:name"`
		Sort int    `gorm:"column:sort"`
	}
	err := dictionaryService.db.WithContext(ctx).
		Table("s_dict AS d").
		Select("d.type, dd.value AS code, dd.label AS name, dd.sort").
		Joins("JOIN s_dict_detail AS dd ON dd.sys_dictionary_id = d.id").
		Where("d.status IS TRUE AND dd.status IS TRUE AND d.deleted_at IS NULL AND dd.deleted_at IS NULL").
		Order("d.type, dd.sort, dd.id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	options := make(map[string][]response.DictionaryOption)
	for _, row := range rows {
		options[row.Type] = append(options[row.Type], response.DictionaryOption{Code: row.Code, Name: row.Name})
	}
	return options, nil
}

// checkCircularReference 沿父链检查 currentID 是否会形成字典环路。
func (dictionaryService *DictionaryService) checkCircularReference(currentID, parentID uint) error {
	if currentID == parentID {
		return errors.New("不能将字典设置为自己的父级")
	}
	var parent model.Dictionary
	err := dictionaryService.db.Where("id = ?", parentID).First(&parent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if parent.ParentID != nil && *parent.ParentID != 0 {
		return dictionaryService.checkCircularReference(currentID, *parent.ParentID)
	}
	return nil
}
