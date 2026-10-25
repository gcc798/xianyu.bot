package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const credentialKeyEnvironment = "XIANYU_CREDENTIAL_KEY"

const credentialKeySetting = "xianyu_credential_key"

func encryptCredential(plaintext, associatedData string) (string, error) {
	key, err := credentialKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), []byte(associatedData))
	return base64.RawStdEncoding.EncodeToString(sealed), nil
}

func decryptCredential(encoded, associatedData string) (string, error) {
	key, err := credentialKey()
	if err != nil {
		return "", err
	}
	sealed, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode credential: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(sealed) < gcm.NonceSize() {
		return "", errors.New("invalid credential ciphertext")
	}
	plaintext, err := gcm.Open(nil, sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():], []byte(associatedData))
	if err != nil {
		return "", fmt.Errorf("decrypt credential: %w", err)
	}
	return string(plaintext), nil
}

func credentialKey() ([]byte, error) {
	encoded := os.Getenv(credentialKeyEnvironment)
	if strings.TrimSpace(encoded) == "" {
		if global.GVA_DB == nil {
			return nil, errors.New("xianyu credential key is not configured: set XIANYU_CREDENTIAL_KEY or initialize the database setting")
		}
		var setting model.Setting
		if err := global.GVA_DB.Where("key = ?", credentialKeySetting).First(&setting).Error; err != nil {
			return nil, fmt.Errorf("load xianyu credential key from database: %w", err)
		}
		encoded = setting.Value
	}
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("%s must be a base64 encoded 32-byte key", credentialKeyEnvironment)
	}
	return key, nil
}

func ValidateCredentialKey() error {
	if _, err := credentialKey(); err != nil {
		return err
	}
	if strings.TrimSpace(os.Getenv(credentialKeyEnvironment)) == "" {
		return nil
	}
	encoded := strings.TrimSpace(os.Getenv(credentialKeyEnvironment))
	db := global.GVA_DB.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	if err := db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&model.Setting{Key: credentialKeySetting, Value: encoded}).Error; err != nil {
		return fmt.Errorf("store xianyu credential key in database: %w", err)
	}
	return nil
}
