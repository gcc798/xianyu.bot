package initialize

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
)

func registerAPIs(apis ...system.SysApi) {
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		for _, api := range apis {
			err := tx.Model(system.SysApi{}).Where("path = ? AND method = ? AND api_group = ? ", api.Path, api.Method, api.ApiGroup).FirstOrCreate(&api).Error
			if err != nil {
				zap.L().Error("注册API失败", zap.Error(err), zap.String("api", api.Path), zap.String("method", api.Method), zap.String("apiGroup", api.ApiGroup))
				return err
			}
		}
		return nil
	})
	if err != nil {
		zap.L().Error("注册API失败", zap.Error(err))
	}
}

func registerMenus(menus ...system.SysBaseMenu) {
	if len(menus) == 0 {
		return
	}
	parentMenu := menus[0]
	parentDefinition := parentMenu
	otherMenus := menus[1:]
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(system.SysBaseMenu{}).Where("name = ? ", parentMenu.Name).FirstOrCreate(&parentMenu).Error
		if err != nil {
			zap.L().Error("注册菜单失败", zap.Error(err))
			return errors.Wrap(err, "注册菜单失败")
		}
		parentDefinition.ID = parentMenu.ID
		if err = updateMenuDefinition(tx, &parentDefinition, parentDefinition.ParentId); err != nil {
			return errors.Wrap(err, "更新根菜单失败")
		}
		pid := parentMenu.ID
		for i := range otherMenus {
			definition := otherMenus[i]
			definition.ParentId = pid
			stored := definition
			stored.Children = nil
			err = tx.Model(system.SysBaseMenu{}).Where("name = ? ", definition.Name).FirstOrCreate(&stored).Error
			if err != nil {
				zap.L().Error("注册菜单失败", zap.Error(err))
				return errors.Wrap(err, "注册菜单失败")
			}
			if len(definition.Children) > 0 {
				definition.ID = stored.ID
				err = updateMenuDefinition(tx, &definition, pid)
				if err != nil {
					return errors.Wrap(err, "更新父菜单失败")
				}
				if err = registerMenuChildren(tx, definition.ID, definition.Children); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		zap.L().Error("注册菜单失败", zap.Error(err))
	}

}

func registerMenuChildren(tx *gorm.DB, parentID uint, menus []system.SysBaseMenu) error {
	for i := range menus {
		definition := menus[i]
		definition.ParentId = parentID
		stored := definition
		stored.Children = nil
		if err := tx.Model(system.SysBaseMenu{}).Where("name = ?", definition.Name).FirstOrCreate(&stored).Error; err != nil {
			return errors.Wrap(err, "注册子菜单失败")
		}
		definition.ID = stored.ID
		if err := updateMenuDefinition(tx, &definition, parentID); err != nil {
			return errors.Wrap(err, "更新子菜单失败")
		}
		if err := registerMenuChildren(tx, definition.ID, definition.Children); err != nil {
			return err
		}
	}
	return nil
}

func updateMenuDefinition(tx *gorm.DB, menu *system.SysBaseMenu, parentID uint) error {
	return tx.Model(menu).Updates(map[string]any{
		"parent_id": parentID,
		"path":      menu.Path,
		"component": menu.Component,
		"sort":      menu.Sort,
		"hidden":    menu.Hidden,
		"title":     menu.Meta.Title,
		"icon":      menu.Meta.Icon,
	}).Error
}
