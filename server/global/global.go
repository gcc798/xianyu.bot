package global

import (
	"github.com/flipped-aurora/gin-vue-admin/server/config"
	"github.com/songzhibin97/gkit/cache/local_cache"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var (
	GVA_DB                  *gorm.DB
	GVA_CONFIG              config.Server
	GVA_LOG                 *zap.Logger
	GVA_Concurrency_Control = &singleflight.Group{}
	BlackCache              local_cache.Cache
)
