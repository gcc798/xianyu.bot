package system

import (
	"errors"
	"strconv"

	"github.com/gcc798/go.xianyu.assistant/server/internal/model/system"
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/system/request"
	"gorm.io/gorm"
)

type DictionaryDetailService struct{ db *gorm.DB }

func (dictionaryDetailService *DictionaryDetailService) CreateSysDictionaryDetail(detail system.SysDictionaryDetail) error {
	if detail.ParentID != nil {
		var parent system.SysDictionaryDetail
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

func (dictionaryDetailService *DictionaryDetailService) DeleteSysDictionaryDetail(detail system.SysDictionaryDetail) error {
	var childCount int64
	if err := dictionaryDetailService.db.Model(&system.SysDictionaryDetail{}).Where("parent_id = ?", detail.ID).Count(&childCount).Error; err != nil {
		return err
	}
	if childCount > 0 {
		return errors.New("该字典详情下还有子项，无法删除")
	}
	return dictionaryDetailService.db.Delete(&detail).Error
}

func (dictionaryDetailService *DictionaryDetailService) UpdateSysDictionaryDetail(detail *system.SysDictionaryDetail) error {
	if detail.ParentID != nil {
		var parent system.SysDictionaryDetail
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
	var parent system.SysDictionaryDetail
	if err := dictionaryDetailService.db.First(&parent, parentID).Error; err != nil || parent.ParentID == nil {
		return false
	}
	return dictionaryDetailService.hasCircularReference(id, *parent.ParentID)
}

func (dictionaryDetailService *DictionaryDetailService) updateChildrenPath(parentID uint) error {
	var parent system.SysDictionaryDetail
	if err := dictionaryDetailService.db.First(&parent, parentID).Error; err != nil {
		return err
	}
	var children []system.SysDictionaryDetail
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

func dictionaryPath(parent system.SysDictionaryDetail) string {
	parentID := strconv.FormatUint(uint64(parent.ID), 10)
	if parent.Path == "" {
		return parentID
	}
	return parent.Path + "," + parentID
}

func (dictionaryDetailService *DictionaryDetailService) GetSysDictionaryDetailInfoList(search request.SysDictionaryDetailSearch) ([]system.SysDictionaryDetail, int64, error) {
	query := dictionaryDetailService.db.Model(&system.SysDictionaryDetail{})
	if search.Label != "" {
		query = query.Where("label LIKE ?", "%"+search.Label+"%")
	}
	if search.Value != "" {
		query = query.Where("value = ?", search.Value)
	}
	if search.Status != nil {
		query = query.Where("status = ?", search.Status)
	}
	if search.SysDictionaryID != 0 {
		query = query.Where("sys_dictionary_id = ?", search.SysDictionaryID)
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
	var details []system.SysDictionaryDetail
	offset := search.PageSize * (search.Page - 1)
	err := query.Limit(search.PageSize).Offset(offset).Order("sort").Order("id").Find(&details).Error
	return details, total, err
}
