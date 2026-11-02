package system

import systemService "github.com/gcc798/go.xianyu.assistant/server/internal/service/system"
import "github.com/gcc798/go.xianyu.assistant/server/internal/security"

type dependencies struct {
	services *systemService.Services
	jwt      *security.JWT
}

type BaseApi struct{ *dependencies }

type API struct {
	JWT              JwtApi
	Base             BaseApi
	Dictionary       DictionaryApi
	DictionaryDetail DictionaryDetailApi
	LoginLog         LoginLogApi
}

func New(services *systemService.Services, jwt *security.JWT) *API {
	deps := &dependencies{services: services, jwt: jwt}
	return &API{
		JWT: JwtApi{deps}, Base: BaseApi{deps}, Dictionary: DictionaryApi{deps},
		DictionaryDetail: DictionaryDetailApi{deps}, LoginLog: LoginLogApi{deps},
	}
}
