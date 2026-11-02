package system

import (
	"context"
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"gorm.io/gorm"
)

type DictionaryService struct{}

var DictionaryServiceApp = new(DictionaryService)

func (dictionaryService *DictionaryService) CreateSysDictionary(dictionary system.SysDictionary) error {
	if !errors.Is(global.GVA_DB.First(&system.SysDictionary{}, "type = ?", dictionary.Type).Error, gorm.ErrRecordNotFound) {
		return errors.New("存在相同的type，不允许创建")
	}
	return global.GVA_DB.Create(&dictionary).Error
}

func (dictionaryService *DictionaryService) DeleteSysDictionary(dictionary system.SysDictionary) error {
	err := global.GVA_DB.Where("id = ?", dictionary.ID).Preload("SysDictionaryDetails").First(&dictionary).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("字典不存在")
	}
	if err != nil {
		return err
	}
	if err := global.GVA_DB.Delete(&dictionary).Error; err != nil {
		return err
	}
	if dictionary.SysDictionaryDetails != nil {
		return global.GVA_DB.Where("sys_dictionary_id = ?", dictionary.ID).Delete(dictionary.SysDictionaryDetails).Error
	}
	return nil
}

func (dictionaryService *DictionaryService) UpdateSysDictionary(dictionary *system.SysDictionary) error {
	var stored system.SysDictionary
	if err := global.GVA_DB.Where("id = ?", dictionary.ID).First(&stored).Error; err != nil {
		return errors.New("查询字典数据失败")
	}
	if stored.Type != dictionary.Type && !errors.Is(global.GVA_DB.First(&system.SysDictionary{}, "type = ?", dictionary.Type).Error, gorm.ErrRecordNotFound) {
		return errors.New("存在相同的type，不允许创建")
	}
	if dictionary.ParentID != nil && *dictionary.ParentID != 0 {
		if err := dictionaryService.checkCircularReference(dictionary.ID, *dictionary.ParentID); err != nil {
			return err
		}
	}
	return global.GVA_DB.Model(&stored).Updates(map[string]any{
		"Name": dictionary.Name, "Type": dictionary.Type, "Status": dictionary.Status,
		"Desc": dictionary.Desc, "ParentID": dictionary.ParentID,
	}).Error
}

func (dictionaryService *DictionaryService) GetSysDictionaryInfoList(ctx context.Context, search request.SysDictionarySearch) ([]system.SysDictionary, error) {
	var dictionaries []system.SysDictionary
	query := global.GVA_DB.WithContext(ctx)
	if search.Name != "" {
		query = query.Where("name LIKE ? OR type LIKE ?", "%"+search.Name+"%", "%"+search.Name+"%")
	}
	err := query.Preload("Children").Find(&dictionaries).Error
	return dictionaries, err
}

func (dictionaryService *DictionaryService) checkCircularReference(currentID, parentID uint) error {
	if currentID == parentID {
		return errors.New("不能将字典设置为自己的父级")
	}
	var parent system.SysDictionary
	err := global.GVA_DB.Where("id = ?", parentID).First(&parent).Error
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
