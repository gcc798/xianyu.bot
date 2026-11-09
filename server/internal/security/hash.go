package security

import "golang.org/x/crypto/bcrypt"

// BcryptHash 使用 bcrypt 默认成本生成不可逆的密码摘要。
func BcryptHash(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes)
}

// BcryptCheck 判断明文密码是否匹配已有 bcrypt 摘要。
func BcryptCheck(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
