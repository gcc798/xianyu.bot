package system

import "github.com/flipped-aurora/gin-vue-admin/server/service"

type BaseApi struct{}

type ApiGroup struct {
	JwtApi
	BaseApi
	DictionaryApi
	DictionaryDetailApi
	LoginLogApi
}

var (
	jwtService              = service.ServiceGroupApp.SystemServiceGroup.JwtService
	userService             = service.ServiceGroupApp.SystemServiceGroup.UserService
	dictionaryService       = service.ServiceGroupApp.SystemServiceGroup.DictionaryService
	dictionaryDetailService = service.ServiceGroupApp.SystemServiceGroup.DictionaryDetailService
	loginLogService         = service.ServiceGroupApp.SystemServiceGroup.LoginLogService
)
