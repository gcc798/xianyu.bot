package utils

import (
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRegisterMenusSupportsNestedChildren(t *testing.T) {
	originalDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = originalDB })

	db, err := gorm.Open(sqlite.Open("file:register-menus?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&system.SysBaseMenu{}))
	global.GVA_DB = db
	require.NoError(t, db.Create(&system.SysBaseMenu{Name: "TestRoot", Path: "root"}).Error)
	require.NoError(t, db.Create(&system.SysBaseMenu{Name: "TestSection", Path: "legacy", Component: "legacy.vue"}).Error)

	RegisterMenus(
		system.SysBaseMenu{Name: "TestRoot", Path: "root"},
		system.SysBaseMenu{
			Name: "TestSection", Path: "section", Component: "view/routerHolder.vue", Meta: system.Meta{Title: "新标题"},
			Children: []system.SysBaseMenu{{Name: "TestChild", Path: "child", Component: "child.vue", Meta: system.Meta{Title: "子页面"}}},
		},
	)

	var root, section, child system.SysBaseMenu
	require.NoError(t, db.Where("name = ?", "TestRoot").First(&root).Error)
	require.NoError(t, db.Where("name = ?", "TestSection").First(&section).Error)
	require.NoError(t, db.Where("name = ?", "TestChild").First(&child).Error)
	require.Equal(t, root.ID, section.ParentId)
	require.Equal(t, section.ID, child.ParentId)
	require.Equal(t, "view/routerHolder.vue", section.Component)
	require.Equal(t, "新标题", section.Meta.Title)
	require.Equal(t, "child.vue", child.Component)
}

func TestRegisterMenusSupportsMultipleRoots(t *testing.T) {
	originalDB := global.GVA_DB
	originalMenus := MenuMap["plugin-tool"]
	t.Cleanup(func() {
		global.GVA_DB = originalDB
		MenuMap["plugin-tool"] = originalMenus
	})

	db, err := gorm.Open(sqlite.Open("file:register-menu-roots?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&system.SysBaseMenu{}))
	global.GVA_DB = db
	MenuMap["plugin-tool"] = nil

	RegisterMenus(
		system.SysBaseMenu{Name: "FirstRoot", Path: "first"},
		system.SysBaseMenu{Name: "FirstChild", Path: "child"},
	)
	RegisterMenus(
		system.SysBaseMenu{Name: "SecondRoot", Path: "second"},
		system.SysBaseMenu{Name: "SecondChild", Path: "child"},
	)

	var firstRoot, secondRoot, secondChild system.SysBaseMenu
	require.NoError(t, db.Where("name = ?", "FirstRoot").First(&firstRoot).Error)
	require.NoError(t, db.Where("name = ?", "SecondRoot").First(&secondRoot).Error)
	require.NoError(t, db.Where("name = ?", "SecondChild").First(&secondChild).Error)
	require.Zero(t, firstRoot.ParentId)
	require.Zero(t, secondRoot.ParentId)
	require.Equal(t, secondRoot.ID, secondChild.ParentId)
	require.Len(t, MenuMap["plugin-tool"], 4)
}
