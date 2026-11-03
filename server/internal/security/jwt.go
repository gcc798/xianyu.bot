package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/gcc798/xianyu.bot/internal/config"
	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/gcc798/xianyu.bot/internal/model/request"
	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/sync/singleflight"
)

// JWT 负责签发、解析和刷新本服务的 HMAC 登录令牌。
type JWT struct {
	// signingKey 是 HMAC 签名密钥。
	signingKey []byte
	// expires 是令牌有效期。
	expires time.Duration
	// buffer 是到期前允许刷新的时间窗口。
	buffer time.Duration
	// issuer 是令牌签发方声明。
	issuer string
}

// NewJWT 校验 settings 中的时长并创建令牌服务。
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

var (
	// TokenValid 表示令牌未通过最终有效性检查。
	TokenValid = errors.New("未知错误")
	// TokenExpired 表示令牌已超过有效期。
	TokenExpired = errors.New("token已过期")
	// TokenNotValidYet 表示令牌尚未到生效时间。
	TokenNotValidYet = errors.New("token尚未激活")
	// TokenMalformed 表示输入不符合 JWT 格式。
	TokenMalformed = errors.New("这不是一个token")
	// TokenSignatureInvalid 表示令牌签名校验失败。
	TokenSignatureInvalid = errors.New("无效签名")
	// TokenInvalid 表示令牌因其他校验错误不可用。
	TokenInvalid = errors.New("无法处理此token")
	// tokenRefresh 合并同一旧令牌触发的并发刷新请求。
	tokenRefresh singleflight.Group
)

// CreateClaims 根据用户基础信息生成带有效期和刷新缓冲期的声明。
func (j *JWT) CreateClaims(baseClaims request.BaseClaims) request.CustomClaims {
	claims := request.CustomClaims{
		BaseClaims: baseClaims,
		BufferTime: int64(j.buffer / time.Second), // 缓冲时间内刷新令牌；前端只保留新令牌。
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"xianyu.bot"},            // 受众
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1000)), // 签名生效时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expires)),
			Issuer:    j.issuer,
		},
	}
	return claims
}

// CreateToken 使用当前 HMAC 密钥签发 claims。
func (j *JWT) CreateToken(claims request.CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.signingKey)
}

// CreateTokenByOldToken 合并同一 oldToken 的并发刷新并签发新令牌。
func (j *JWT) CreateTokenByOldToken(oldToken string, claims request.CustomClaims) (string, error) {
	value, err, _ := tokenRefresh.Do("JWT:"+oldToken, func() (any, error) {
		return j.CreateToken(claims)
	})
	if err != nil {
		return "", err
	}
	return value.(string), nil
}

// ParseToken 验证 tokenString 并返回业务声明及稳定的领域错误。
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

// ExpiresIn 返回令牌有效期。
func (j *JWT) ExpiresIn() time.Duration { return j.expires }

// LoginToken 为 user 创建登录声明并签发令牌。
func (j *JWT) LoginToken(user *model.User) (string, request.CustomClaims, error) {
	claims := j.CreateClaims(request.BaseClaims{
		UUID: user.UUID, ID: user.ID, NickName: user.NickName,
		Username: user.Username, AuthorityID: user.AuthorityID,
	})
	token, err := j.CreateToken(claims)
	return token, claims, err
}
