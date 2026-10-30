package utils

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
)

var (
	ApiMap  = make(map[string][]system.SysApi)
	MenuMap = make(map[string][]system.SysBaseMenu)
	DictMap = make(map[string][]system.SysDictionary)
	rw      sync.Mutex
)

func getPluginName() string {
	_, file, _, ok := runtime.Caller(2)
	pluginName := ""
	if ok {
		file = filepath.ToSlash(file)
		const key = "server/plugin/"
		if idx := strings.Index(file, key); idx != -1 {
			remain := file[idx+len(key):]
			parts := strings.Split(remain, "/")
			if len(parts) > 0 {
				pluginName = parts[0]
			}
		}
	}
	return pluginName
}

func RegisterApis(apis ...system.SysApi) {
	name := getPluginName()
	if name != "" {
		rw.Lock()
		ApiMap[name] = apis
		rw.Unlock()
	}

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

func RegisterMenus(menus ...system.SysBaseMenu) {
	if len(menus) == 0 {
		return
	}
	name := getPluginName()
	if name != "" {
		rw.Lock()
		MenuMap[name] = append(MenuMap[name], flattenMenus(menus)...)
		rw.Unlock()
	}

	parentMenu := menus[0]
	otherMenus := menus[1:]
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(system.SysBaseMenu{}).Where("name = ? ", parentMenu.Name).FirstOrCreate(&parentMenu).Error
		if err != nil {
			zap.L().Error("注册菜单失败", zap.Error(err))
			return errors.Wrap(err, "注册菜单失败")
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

func flattenMenus(menus []system.SysBaseMenu) []system.SysBaseMenu {
	flattened := make([]system.SysBaseMenu, 0, len(menus))
	var appendMenus func([]system.SysBaseMenu)
	appendMenus = func(entries []system.SysBaseMenu) {
		for _, menu := range entries {
			children := menu.Children
			menu.Children = nil
			flattened = append(flattened, menu)
			appendMenus(children)
		}
	}
	appendMenus(menus)
	return flattened
}

func RegisterDictionaries(dictionaries ...system.SysDictionary) {
	name := getPluginName()
	if name != "" {
		rw.Lock()
		DictMap[name] = dictionaries
		rw.Unlock()
	}

	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		for _, dict := range dictionaries {
			details := dict.SysDictionaryDetails
			dict.SysDictionaryDetails = nil
			err := tx.Model(system.SysDictionary{}).Where("type = ?", dict.Type).FirstOrCreate(&dict).Error
			if err != nil {
				zap.L().Error("注册字典失败", zap.Error(err), zap.String("type", dict.Type))
				return err
			}
			for _, detail := range details {
				detail.SysDictionaryID = int(dict.ID)
				err = tx.Model(system.SysDictionaryDetail{}).Where("sys_dictionary_id = ? AND value = ?", dict.ID, detail.Value).FirstOrCreate(&detail).Error
				if err != nil {
					zap.L().Error("注册字典详情失败", zap.Error(err), zap.String("value", detail.Value))
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		zap.L().Error("注册字典失败", zap.Error(err))
	}
}

func Pointer[T any](in T) *T {
	return &in
}

func GetPluginData(pluginName string) ([]system.SysApi, []system.SysBaseMenu, []system.SysDictionary) {
	rw.Lock()
	defer rw.Unlock()
	return ApiMap[pluginName], MenuMap[pluginName], DictMap[pluginName]
}
