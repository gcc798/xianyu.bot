package system

import systemAPI "github.com/gcc798/go.xianyu.assistant/server/internal/api/system"

type dependencies struct{ handlers *systemAPI.API }

type RouterGroup struct {
	JwtRouter
	BaseRouter
	UserRouter
	DictionaryRouter
	DictionaryDetailRouter
	LoginLogRouter
}

func New(handlers *systemAPI.API) *RouterGroup {
	deps := &dependencies{handlers: handlers}
	return &RouterGroup{
		JwtRouter: JwtRouter{deps}, BaseRouter: BaseRouter{deps}, UserRouter: UserRouter{deps},
		DictionaryRouter: DictionaryRouter{deps}, DictionaryDetailRouter: DictionaryDetailRouter{deps},
		LoginLogRouter: LoginLogRouter{deps},
	}
}
