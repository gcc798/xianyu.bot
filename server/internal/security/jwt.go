package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/gcc798/xianyu.bot/internal/config"
	"github.com/gcc798/xianyu.bot/internal/model/system"
	"github.com/gcc798/xianyu.bot/internal/model/system/request"
	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/sync/singleflight"
)

type JWT struct {
	signingKey []byte
	expires    time.Duration
	buffer     time.Duration
	issuer     string
}

var (
	TokenValid            = errors.New("未知错误")
	TokenExpired          = errors.New("token已过期")
	TokenNotValidYet      = errors.New("token尚未激活")
	TokenMalformed        = errors.New("这不是一个token")
	TokenSignatureInvalid = errors.New("无效签名")
	TokenInvalid          = errors.New("无法处理此token")
	tokenRefresh          singleflight.Group
)

func NewJWT(settings config.JWT) (*JWT, error) {
	expires, err := ParseDuration(settings.ExpiresTime)
	if err != nil {
		return nil, fmt.Errorf("expires-time: %w", err)
	}
	buffer, err := ParseDuration(settings.BufferTime)
	if err != nil {
		return nil, fmt.Errorf("buffer-time: %w", err)
	}
	return &JWT{signingKey: []byte(settings.SigningKey), expires: expires, buffer: buffer, issuer: settings.Issuer}, nil
}

func (j *JWT) CreateClaims(baseClaims request.BaseClaims) request.CustomClaims {
	claims := request.CustomClaims{
		BaseClaims: baseClaims,
		BufferTime: int64(j.buffer / time.Second), // 缓冲时间内刷新令牌；前端只保留新令牌。
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"GVA"},                   // 受众
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1000)), // 签名生效时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expires)),
			Issuer:    j.issuer,
		},
	}
	return claims
}

// CreateToken 创建一个token
func (j *JWT) CreateToken(claims request.CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.signingKey)
}

// CreateTokenByOldToken 旧token 换新token 使用归并回源避免并发问题
func (j *JWT) CreateTokenByOldToken(oldToken string, claims request.CustomClaims) (string, error) {
	value, err, _ := tokenRefresh.Do("JWT:"+oldToken, func() (any, error) {
		return j.CreateToken(claims)
	})
	if err != nil {
		return "", err
	}
	return value.(string), nil
}

// ParseToken 解析 token
func (j *JWT) ParseToken(tokenString string) (*request.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &request.CustomClaims{}, func(token *jwt.Token) (any, error) {
		return j.signingKey, nil
	})

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, TokenExpired
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, TokenMalformed
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, TokenSignatureInvalid
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, TokenNotValidYet
		default:
			return nil, TokenInvalid
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*request.CustomClaims); ok && token.Valid {
			return claims, nil
		}
	}
	return nil, TokenValid
}

func (j *JWT) ExpiresIn() time.Duration { return j.expires }

func (j *JWT) LoginToken(user *system.User) (string, request.CustomClaims, error) {
	claims := j.CreateClaims(request.BaseClaims{
		UUID: user.UUID, ID: user.ID, NickName: user.NickName,
		Username: user.Username, AuthorityID: user.AuthorityID,
	})
	token, err := j.CreateToken(claims)
	return token, claims, err
}
