package system

import (
	"errors"
	"strconv"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/gcc798/xianyu.bot/internal/model/request"
	"gorm.io/gorm"
)

type DictionaryDetailService struct{ db *gorm.DB }

func (dictionaryDetailService *DictionaryDetailService) Create(detail model.DictionaryDetail) error {
	if detail.ParentID != nil {
		var parent model.DictionaryDetail
		if err := dictionaryDetailService.db.First(&parent, *detail.ParentID).Error; err != nil {
			return err
		}
		detail.Level = parent.Level + 1
		detail.Path = dictionaryPath(parent)
	} else {
		detail.Level = 0
		detail.Path = ""
	}
	return dictionaryDetailService.db.Create(&detail).Error
}

func (dictionaryDetailService *DictionaryDetailService) Delete(detail model.DictionaryDetail) error {
	var childCount int64
	if err := dictionaryDetailService.db.Model(&model.DictionaryDetail{}).Where("parent_id = ?", detail.ID).Count(&childCount).Error; err != nil {
		return err
	}
	if childCount > 0 {
		return errors.New("该字典详情下还有子项，无法删除")
	}
	return dictionaryDetailService.db.Delete(&detail).Error
}

func (dictionaryDetailService *DictionaryDetailService) Update(detail *model.DictionaryDetail) error {
	if detail.ParentID != nil {
		var parent model.DictionaryDetail
		if err := dictionaryDetailService.db.First(&parent, *detail.ParentID).Error; err != nil {
			return err
		}
		if dictionaryDetailService.hasCircularReference(detail.ID, *detail.ParentID) {
			return errors.New("不能将字典详情设置为自己或其子项的父级")
		}
		detail.Level = parent.Level + 1
		detail.Path = dictionaryPath(parent)
	} else {
		detail.Level = 0
		detail.Path = ""
	}
	if err := dictionaryDetailService.db.Save(detail).Error; err != nil {
		return err
	}
	return dictionaryDetailService.updateChildrenPath(detail.ID)
}

func (dictionaryDetailService *DictionaryDetailService) hasCircularReference(id, parentID uint) bool {
	if id == parentID {
		return true
	}
	var parent model.DictionaryDetail
	if err := dictionaryDetailService.db.First(&parent, parentID).Error; err != nil || parent.ParentID == nil {
		return false
	}
	return dictionaryDetailService.hasCircularReference(id, *parent.ParentID)
}

func (dictionaryDetailService *DictionaryDetailService) updateChildrenPath(parentID uint) error {
	var parent model.DictionaryDetail
	if err := dictionaryDetailService.db.First(&parent, parentID).Error; err != nil {
		return err
	}
	var children []model.DictionaryDetail
	if err := dictionaryDetailService.db.Where("parent_id = ?", parentID).Find(&children).Error; err != nil {
		return err
	}
	for i := range children {
		children[i].Level = parent.Level + 1
		children[i].Path = dictionaryPath(parent)
		if err := dictionaryDetailService.db.Save(&children[i]).Error; err != nil {
			return err
		}
		if err := dictionaryDetailService.updateChildrenPath(children[i].ID); err != nil {
			return err
		}
	}
	return nil
}

func dictionaryPath(parent model.DictionaryDetail) string {
	parentID := strconv.FormatUint(uint64(parent.ID), 10)
	if parent.Path == "" {
		return parentID
	}
	return parent.Path + "," + parentID
}

func (dictionaryDetailService *DictionaryDetailService) List(search request.DictionaryDetailSearch) ([]model.DictionaryDetail, int64, error) {
	query := dictionaryDetailService.db.Model(&model.DictionaryDetail{})
	if search.Label != "" {
		query = query.Where("label LIKE ?", "%"+search.Label+"%")
	}
	if search.Value != "" {
		query = query.Where("value = ?", search.Value)
	}
	if search.Status != nil {
		query = query.Where("status = ?", search.Status)
	}
	if search.DictionaryID != 0 {
		query = query.Where("sys_dictionary_id = ?", search.DictionaryID)
	}
	if search.ParentID != nil {
		query = query.Where("parent_id = ?", *search.ParentID)
	}
	if search.Level != 0 {
		query = query.Where("level = ?", search.Level)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var details []model.DictionaryDetail
	offset := search.PageSize * (search.Page - 1)
	err := query.Limit(search.PageSize).Offset(offset).Order("sort").Order("id").Find(&details).Error
	return details, total, err
}
