package system

import "gorm.io/gorm"

type Services struct {
	RevokedTokens    RevokedTokenService
	User             UserService
	Dictionary       DictionaryService
	DictionaryDetail DictionaryDetailService
	LoginLog         LoginLogService
}

func New(db *gorm.DB) *Services {
	return &Services{
		RevokedTokens: RevokedTokenService{db: db}, User: UserService{db: db},
		Dictionary: DictionaryService{db: db}, DictionaryDetail: DictionaryDetailService{db: db},
		LoginLog: LoginLogService{db: db},
	}
}
