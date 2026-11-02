package system

import (
	"context"
	"errors"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/gcc798/xianyu.bot/internal/model/request"
	"gorm.io/gorm"
)

type DictionaryService struct{ db *gorm.DB }

func (dictionaryService *DictionaryService) Create(dictionary model.Dictionary) error {
	if !errors.Is(dictionaryService.db.First(&model.Dictionary{}, "type = ?", dictionary.Type).Error, gorm.ErrRecordNotFound) {
		return errors.New("存在相同的type，不允许创建")
	}
	return dictionaryService.db.Create(&dictionary).Error
}

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

func (dictionaryService *DictionaryService) List(ctx context.Context, search request.DictionarySearch) ([]model.Dictionary, error) {
	var dictionaries []model.Dictionary
	query := dictionaryService.db.WithContext(ctx)
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
