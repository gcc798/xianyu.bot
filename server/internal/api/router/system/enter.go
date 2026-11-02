package system

import systemAPI "github.com/gcc798/xianyu.bot/internal/api/system"

type dependencies struct{ handlers *systemAPI.API }

type RouterGroup struct {
	JWTRouter
	AuthRouter
	UserRouter
	DictionaryRouter
	DictionaryDetailRouter
	LoginLogRouter
}

func New(handlers *systemAPI.API) *RouterGroup {
	deps := &dependencies{handlers: handlers}
	return &RouterGroup{
		JWTRouter: JWTRouter{deps}, AuthRouter: AuthRouter{deps}, UserRouter: UserRouter{deps},
		DictionaryRouter: DictionaryRouter{deps}, DictionaryDetailRouter: DictionaryDetailRouter{deps},
		LoginLogRouter: LoginLogRouter{deps},
	}
}
