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

	"github.com/gcc798/xianyu.bot/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// credentialKeyEnvironment 指定可覆盖数据库密钥的环境变量。
const credentialKeyEnvironment = "XY_CREDENTIAL_KEY"

// credentialKeySetting 是数据库中保存凭据密钥的配置键。
const credentialKeySetting = "xy_credential_key"

// encryptCredential 使用 AES-GCM 和 associatedData 加密敏感文本。
func encryptCredential(db *gorm.DB, plaintext, associatedData string) (string, error) {
	key, err := credentialKey(db)
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

// decryptCredential 使用同一 associatedData 验证并解密凭据密文。
func decryptCredential(db *gorm.DB, encoded, associatedData string) (string, error) {
	key, err := credentialKey(db)
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

// credentialKey 优先读取环境变量，否则读取数据库中的 32 字节密钥。
func credentialKey(db *gorm.DB) ([]byte, error) {
	encoded := os.Getenv(credentialKeyEnvironment)
	if strings.TrimSpace(encoded) == "" {
		if db == nil {
			return nil, errors.New("xy credential key is not configured: set XY_CREDENTIAL_KEY or initialize the database setting")
		}
		var setting model.Setting
		if err := db.Where("key = ?", credentialKeySetting).First(&setting).Error; err != nil {
			return nil, fmt.Errorf("load xy credential key from database: %w", err)
		}
		encoded = setting.Value
	}
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("%s must be a base64 encoded 32-byte key", credentialKeyEnvironment)
	}
	return key, nil
}

// ValidateCredentialKey 校验或首次生成凭据密钥，，环境密钥仅在进程内使用。
func ValidateCredentialKey(db *gorm.DB) error {
	encoded := strings.TrimSpace(os.Getenv(credentialKeyEnvironment))
	if encoded != "" {
		_, err := credentialKey(nil)
		return err
	}
	if _, err := credentialKey(db); err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("generate xy credential key: %w", err)
	}
	encoded = base64.StdEncoding.EncodeToString(key)
	db = db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	conflict := clause.OnConflict{DoNothing: true}
	if err := db.Clauses(conflict).Create(&model.Setting{Key: credentialKeySetting, Value: encoded}).Error; err != nil {
		return fmt.Errorf("store xy credential key in database: %w", err)
	}
	_, err := credentialKey(db)
	return err
}
