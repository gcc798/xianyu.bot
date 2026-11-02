package system

import api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"

type RouterGroup struct {
	JwtRouter
	BaseRouter
	UserRouter
	DictionaryRouter
	DictionaryDetailRouter
	LoginLogRouter
}

var (
	jwtApi              = api.ApiGroupApp.SystemApiGroup.JwtApi
	baseApi             = api.ApiGroupApp.SystemApiGroup.BaseApi
	dictionaryApi       = api.ApiGroupApp.SystemApiGroup.DictionaryApi
	dictionaryDetailApi = api.ApiGroupApp.SystemApiGroup.DictionaryDetailApi
)
