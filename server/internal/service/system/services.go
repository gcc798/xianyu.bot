package system

import "gorm.io/gorm"

type Services struct {
	JWT              JwtService
	User             UserService
	Dictionary       DictionaryService
	DictionaryDetail DictionaryDetailService
	LoginLog         LoginLogService
}

func New(db *gorm.DB) *Services {
	return &Services{
		JWT: JwtService{db: db}, User: UserService{db: db},
		Dictionary: DictionaryService{db: db}, DictionaryDetail: DictionaryDetailService{db: db},
		LoginLog: LoginLogService{db: db},
	}
}
