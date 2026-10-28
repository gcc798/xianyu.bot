package main

import (
	"cmp"
	"fmt"
	"os"

	"github.com/flipped-aurora/gin-vue-admin/server/core"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"go.uber.org/zap"
)

func main() {
	global.GVA_VP = core.Viper()
	initialize.OtherInit()
	global.GVA_LOG = core.Zap()
	zap.ReplaceGlobals(global.GVA_LOG)
	global.GVA_DB = initialize.Gorm()

	password := cmp.Or(os.Getenv("GVA_ADMIN_PASSWORD"), "123456")
	if err := new(system.InitDBService).InitExistingPgsql(global.GVA_DB, password); err != nil {
		global.GVA_LOG.Fatal("official database initialization failed", zap.Error(err))
	}
	fmt.Println("gin-vue-admin official database initialization completed")
}
