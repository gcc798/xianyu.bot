package system

import (
	"github.com/gcc798/xianyu.bot/internal/security"
	systemService "github.com/gcc798/xianyu.bot/internal/service/system"
)

type dependencies struct {
	services *systemService.Services
	jwt      *security.JWT
}

type UserAPI struct{ *dependencies }

type API struct {
	JWT              JWTAPI
	User             UserAPI
	Dictionary       DictionaryAPI
	DictionaryDetail DictionaryDetailAPI
	LoginLog         LoginLogAPI
}

func New(services *systemService.Services, jwt *security.JWT) *API {
	deps := &dependencies{services: services, jwt: jwt}
	return &API{
		JWT: JWTAPI{deps}, User: UserAPI{deps}, Dictionary: DictionaryAPI{deps},
		DictionaryDetail: DictionaryDetailAPI{deps}, LoginLog: LoginLogAPI{deps},
	}
}
