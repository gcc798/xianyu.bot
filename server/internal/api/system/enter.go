package system

import (
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gcc798/xianyu.bot/internal/service"
)

type dependencies struct {
	services *service.Service
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

func New(services *service.Service, jwt *security.JWT) *API {
	deps := &dependencies{services: services, jwt: jwt}
	return &API{
		JWT: JWTAPI{deps}, User: UserAPI{deps}, Dictionary: DictionaryAPI{deps},
		DictionaryDetail: DictionaryDetailAPI{deps}, LoginLog: LoginLogAPI{deps},
	}
}
