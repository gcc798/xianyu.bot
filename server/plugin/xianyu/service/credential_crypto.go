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
)

const credentialKeyEnvironment = "XIANYU_CREDENTIAL_KEY"

// This file is a local-development fallback; deployments should set the environment variable.
const localCredentialKeyFile = "/Users/guoc/business/xinayu/go.xianyu.assistant.secret.txt"

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
		contents, err := os.ReadFile(localCredentialKeyFile)
		if err != nil {
			return nil, fmt.Errorf("read local credential key: %w", err)
		}
		encoded = strings.TrimSpace(string(contents))
	}
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("%s must be a base64 encoded 32-byte key", credentialKeyEnvironment)
	}
	return key, nil
}
