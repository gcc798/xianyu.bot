package initialize

import (
	"context"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"gorm.io/gorm"
)

const (
	legacyRoutePrefix    = "/xianyu/"
	currentRoutePrefix   = "/xy/"
	legacyComponentPath  = "plugin/xianyu/"
	currentComponentPath = "plugin/xy/"
)

// Naming migrates persisted menu and API identifiers while preserving their IDs and authority relations.
func Naming(ctx context.Context) error {
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := migrateAPIPaths(tx); err != nil {
			return err
		}
		return migrateMenus(tx)
	})
}

func migrateAPIPaths(tx *gorm.DB) error {
	var apis []system.SysApi
	if err := tx.Where("path LIKE ?", legacyRoutePrefix+"%").Find(&apis).Error; err != nil {
		return err
	}
	for i := range apis {
		path, _ := strings.CutPrefix(apis[i].Path, legacyRoutePrefix)
		if err := tx.Model(&apis[i]).Update("path", currentRoutePrefix+path).Error; err != nil {
			return err
		}
	}
	return nil
}

func migrateMenus(tx *gorm.DB) error {
	var menus []system.SysBaseMenu
	if err := tx.Where("name LIKE ? OR component LIKE ? OR path = ?", "Xianyu%", legacyComponentPath+"%", "xianyu").Find(&menus).Error; err != nil {
		return err
	}
	for i := range menus {
		updates := map[string]any{}
		if name, found := strings.CutPrefix(menus[i].Name, "Xianyu"); found {
			updates["name"] = "Xy" + name
		}
		if component, found := strings.CutPrefix(menus[i].Component, legacyComponentPath); found {
			updates["component"] = currentComponentPath + component
		}
		if menus[i].Path == "xianyu" {
			updates["path"] = "xy"
		}
		if len(updates) > 0 {
			if err := tx.Model(&menus[i]).Updates(updates).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
