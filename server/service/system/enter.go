package system

type ServiceGroup struct {
	JwtService
	UserService
	DictionaryService
	DictionaryDetailService
	LoginLogService
}
