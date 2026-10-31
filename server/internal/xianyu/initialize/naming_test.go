package initialize

import (
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNamingMigratesPersistedAPIAndMenuIdentifiers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&system.SysApi{}, &system.SysBaseMenu{}))

	api := system.SysApi{Path: "/xianyu/status", Method: "GET", ApiGroup: "闲鱼助手"}
	menu := system.SysBaseMenu{Path: "xianyu", Name: "Xianyu", Component: "plugin/xianyu/view/dashboard.vue"}
	require.NoError(t, db.Create(&api).Error)
	require.NoError(t, db.Create(&menu).Error)

	oldDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = oldDB })
	global.GVA_DB = db

	require.NoError(t, Naming(t.Context()))
	require.NoError(t, Naming(t.Context()))

	require.NoError(t, db.First(&api, api.ID).Error)
	require.Equal(t, "/xy/status", api.Path)
	require.NoError(t, db.First(&menu, menu.ID).Error)
	require.Equal(t, "xy", menu.Path)
	require.Equal(t, "Xy", menu.Name)
	require.Equal(t, "view/xianyu/dashboard.vue", menu.Component)
}

func TestNamingMigratesPluginComponentPath(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&system.SysApi{}, &system.SysBaseMenu{}))
	menu := system.SysBaseMenu{Name: "XyDashboard", Component: "plugin/xy/view/dashboard.vue"}
	require.NoError(t, db.Create(&menu).Error)

	oldDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = oldDB })
	global.GVA_DB = db

	require.NoError(t, Naming(t.Context()))
	require.NoError(t, db.First(&menu, menu.ID).Error)
	require.Equal(t, "view/xianyu/dashboard.vue", menu.Component)
}
